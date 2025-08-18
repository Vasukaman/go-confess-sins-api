package ttsservice

import "context"

// TTSClient is an interface that describes a text-to-speech client.
type TTSClient interface {
	SynthesizeSpeech(ctx context.Context, text string) ([]byte, error)
	Close()
}
