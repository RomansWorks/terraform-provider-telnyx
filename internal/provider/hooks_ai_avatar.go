// Copyright (c) 2026 Roman Baron
// SPDX-License-Identifier: MIT

package provider

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
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

// maxAvatarBytes is the API's upload limit.
const maxAvatarBytes = 10 << 20

// Avatars are created by uploading a file. The file's hash is tracked so a
// new image replaces the avatar; the API cannot change an avatar's image.
var aiAvatarHooks = engine.Hooks{
	ModifyPlan: func(_ context.Context, plan, state engine.Values) ([]string, diag.Diagnostics) {
		var diags diag.Diagnostics
		p := plan["image_path"]
		if !p.IsKnown() || p.IsNull() {
			plan["image_sha256"] = tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
			return nil, nil
		}
		var path string
		_ = p.As(&path)
		sum, err := fileSHA256(path)
		if err != nil {
			diags.AddAttributeError(pathRoot("image_path"), "Unable to read avatar image", err.Error())
			return nil, diags
		}
		plan["image_sha256"] = tftypes.NewValue(tftypes.String, sum)
		if old, ok := state["image_sha256"]; ok && !old.IsNull() && !old.Equal(plan["image_sha256"]) {
			return []string{"image_sha256"}, nil
		}
		return nil, nil
	},
	Create: func(ctx context.Context, c *client.Client, r *engine.Resource, plan engine.Values) (engine.Object, error) {
		var name, path string
		if err := plan["name"].As(&name); err != nil {
			return nil, err
		}
		if err := plan["image_path"].As(&path); err != nil {
			return nil, err
		}
		body, contentType, err := avatarForm(name, path)
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

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

var imageTypes = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"}

// avatarForm builds the multipart body of an avatar upload.
func avatarForm(name, path string) ([]byte, string, error) {
	ct, ok := imageTypes[strings.ToLower(filepath.Ext(path))]
	if !ok {
		return nil, "", fmt.Errorf("%s: the image must be a .jpg, .jpeg, .png or .webp file", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	if len(data) > maxAvatarBytes {
		return nil, "", fmt.Errorf("%s is %d bytes; the limit is 10 MiB", path, len(data))
	}
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if err := w.WriteField("name", name); err != nil {
		return nil, "", err
	}
	h := textproto.MIMEHeader{}
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image_file"; filename=%q`, filepath.Base(path)))
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
