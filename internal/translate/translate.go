package translate

import "context"

type Translator interface {
	Translate(ctx context.Context, text, from, to, format string) (string, error)
}
