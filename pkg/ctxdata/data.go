package ctxdata

import "context"

func GetUId(ctx context.Context) string {
	if u, ok := ctx.Value("uid").(string); ok {
		return u
	}
	return ""
}
