package google

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

type Client struct {
	ttsClient *texttospeech.Client
}

// NewClient creates a new client using the downloaded JSON key file.
func NewClient(ctx context.Context, credentialsJSON []byte) (*Client, error) {
	client, err := texttospeech.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client from JSON: %w", err)
	}
	return &Client{ttsClient: client}, nil
}

// SynthesizeSpeech takes text and returns the MP3 audio data as a byte slice.
func (c *Client) SynthesizeSpeech(ctx context.Context, text string) ([]byte, error) {
	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   texttospeechpb.SsmlVoiceGender_NEUTRAL,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  1.15, // A little sped up, as you wanted
		},
	}

	resp, err := c.ttsClient.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SynthesizeSpeech failed: %w", err)
	}

	return resp.AudioContent, nil
}

func (c *Client) Close() {
	c.ttsClient.Close()
}
