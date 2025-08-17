package google

import (
	"context"
	"fmt"
	"math/rand" // Import the rand package
	"time"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"google.golang.org/api/option"
)

// Define our voice options
var genders = []texttospeechpb.SsmlVoiceGender{
	texttospeechpb.SsmlVoiceGender_NEUTRAL,
	texttospeechpb.SsmlVoiceGender_MALE,
	texttospeechpb.SsmlVoiceGender_FEMALE,
}

const minSpeakingRate = 0.85
const maxSpeakingRate = 1.4

type Client struct {
	ttsClient *texttospeech.Client
	rng       *rand.Rand // Add a random number generator
}

func NewClient(ctx context.Context, credentialsJSON []byte) (*Client, error) {
	client, err := texttospeech.NewClient(ctx, option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create TTS client from JSON: %w", err)
	}
	// Seed the random number generator
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return &Client{ttsClient: client, rng: rng}, nil
}

// SynthesizeSpeech now uses random settings.
func (c *Client) SynthesizeSpeech(ctx context.Context, text string) ([]byte, error) {
	// 1. Pick a random gender from our list.
	randomGender := genders[c.rng.Intn(len(genders))]

	// 2. Pick a random speaking rate in our desired range.
	randomRate := minSpeakingRate + c.rng.Float64()*(maxSpeakingRate-minSpeakingRate)

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-US",
			SsmlGender:   randomGender, // Use the random gender
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
			SpeakingRate:  randomRate, // Use the random rate
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
