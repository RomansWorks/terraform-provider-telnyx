// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/RomansWorks/terraform-provider-telnyx/internal/client"
	"github.com/RomansWorks/terraform-provider-telnyx/internal/engine"
)

// maxVoiceSampleBytes is the API's upload limit for a voice recording.
const maxVoiceSampleBytes = 20 << 20

var audioTypes = map[string]string{
	".wav": "audio/wav", ".mp3": "audio/mpeg", ".flac": "audio/flac", ".ogg": "audio/ogg", ".m4a": "audio/mp4",
}

// voiceCloneFromDesignPath is the create of a clone from a voice design;
// the descriptor's create is the upload of a recording.
const voiceCloneFromDesignPath = "/voice_clones"

// voiceCloneDesignFields are the fields the create from a design takes.
var voiceCloneDesignFields = []string{"name", "voice_design_id", "language", "gender", "voice_provider"}

// A voice clone is made from a voice design with a JSON request, or from a
// recording by uploading the file with the clone's settings as a multipart
// form. The file's hash is tracked so a new recording replaces the clone;
// the API cannot change a clone's recording.
var voiceCloneHooks = engine.Hooks{
	ModifyPlan: func(_ context.Context, plan, state engine.Values) ([]string, diag.Diagnostics) {
		var diags diag.Diagnostics
		// The create from a design takes none of the recording's settings,
		// and the API fills label from the design's prompt.
		if v := plan["voice_design_id"]; !v.IsNull() {
			for _, name := range []string{"label", "model_id", "ref_text"} {
				if x, ok := plan[name]; ok && x.IsKnown() && !x.IsNull() && !stateHas(state, name, x) {
					diags.AddAttributeError(pathRoot(name), "Not allowed with voice_design_id",
						name+" applies only to a clone made from a recording (audio_path).")
				}
			}
			if diags.HasError() {
				return nil, diags
			}
		}
		p := plan["audio_path"]
		if !p.IsKnown() {
			plan["audio_sha256"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			return nil, nil
		}
		if p.IsNull() {
			plan["audio_sha256"] = tftypes.NewValue(tftypes.String, nil)
			return nil, nil
		}
		var path string
		_ = p.As(&path)
		_, err := audioType(path)
		sum := ""
		if err == nil {
			sum, err = fileSHA256(path)
		}
		if err != nil {
			// An existing clone keeps its recording when the file is not on
			// this machine, such as a plan run elsewhere.
			if old, ok := state["audio_sha256"]; ok && !old.IsNull() && errors.Is(err, fs.ErrNotExist) {
				plan["audio_sha256"] = old
				diags.AddAttributeWarning(pathRoot("audio_path"), "Voice recording not found",
					"The recording was not found, so a change to it cannot be detected: "+err.Error())
				return nil, diags
			}
			diags.AddAttributeError(pathRoot("audio_path"), "Unable to read voice recording", err.Error())
			return nil, diags
		}
		plan["audio_sha256"] = tftypes.NewValue(tftypes.String, sum)
		if old, ok := state["audio_sha256"]; ok && !old.IsNull() && !old.Equal(plan["audio_sha256"]) {
			return []string{"audio_sha256"}, nil
		}
		return nil, nil
	},
	// The update request requires name, but the engine sends only what
	// changed.
	BeforeRequest: func(_ context.Context, op string, body engine.Object, plan engine.Values) error {
		var name string
		if op == "update" && plan["name"].IsKnown() && plan["name"].As(&name) == nil && name != "" {
			body["name"] = name
		}
		return nil
	},
	Create: func(ctx context.Context, c *client.Client, r *engine.Resource, plan engine.Values) (engine.Object, error) {
		if v := plan["voice_design_id"]; v.IsKnown() && !v.IsNull() {
			body := map[string]any{}
			for _, name := range voiceCloneDesignFields {
				a := r.Attr(name)
				var s string
				if x := plan[name]; a != nil && x.IsKnown() && !x.IsNull() && x.As(&s) == nil {
					key := a.JSON
					if key == "" {
						key = a.Name
					}
					body[key] = s
				}
			}
			data, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			out, err := c.Send(ctx, http.MethodPost, voiceCloneFromDesignPath, nil, "application/json", data)
			if err != nil {
				return nil, err
			}
			return engine.Decode(out, true)
		}
		body, contentType, err := voiceCloneForm(r, plan)
		if err != nil {
			return nil, err
		}
		out, err := c.Send(ctx, http.MethodPost, r.Create.Path, nil, contentType, body)
		if err != nil {
			return nil, err
		}
		return engine.Decode(out, r.Create.Envelope)
	},
}

func audioType(path string) (string, error) {
	ct, ok := audioTypes[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return "", fmt.Errorf("%s: the recording must be a .wav, .mp3, .flac, .ogg or .m4a file", path)
	}
	return ct, nil
}

// voiceCloneForm builds the multipart body of a voice clone upload: every
// string field the create request takes, then the recording.
func voiceCloneForm(r *engine.Resource, plan engine.Values) ([]byte, string, error) {
	var path string
	if err := plan["audio_path"].As(&path); err != nil {
		return nil, "", err
	}
	ct, err := audioType(path)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxVoiceSampleBytes {
		return nil, "", fmt.Errorf("%s is %d bytes; the limit is 20 MiB", path, len(data))
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for _, a := range r.Attrs {
		v := plan[a.Name]
		if !a.InCreate || a.Local || a.Kind != engine.KindString || !v.IsKnown() || v.IsNull() {
			continue
		}
		var s string
		if err := v.As(&s); err != nil {
			return nil, "", err
		}
		if err := w.WriteField(a.JSON, s); err != nil {
			return nil, "", err
		}
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="audio_file"; filename=%q`, filepath.Base(path)))
	h.Set("Content-Type", ct)
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, "", err
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), w.FormDataContentType(), nil
}

// stateHas says the value is the one already in state, which the API set
// itself (such as label, filled from the design's prompt).
func stateHas(state engine.Values, name string, v tftypes.Value) bool {
	old, ok := state[name]
	return ok && old.Equal(v)
}
