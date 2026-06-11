package main

// Built-in sacrifice registry: dollars per million tokens.
// Override stale entries (or add new models) via the `pricing:`
// map in openpray.yaml.
var builtinPricing = map[string]Pricing{
	// Anthropic
	"claude-fable-5":    {InputPerMTok: 10.00, OutputPerMTok: 50.00},
	"claude-opus-4-8":   {InputPerMTok: 5.00, OutputPerMTok: 25.00},
	"claude-opus-4-7":   {InputPerMTok: 5.00, OutputPerMTok: 25.00},
	"claude-opus-4-6":   {InputPerMTok: 5.00, OutputPerMTok: 25.00},
	"claude-sonnet-4-6": {InputPerMTok: 3.00, OutputPerMTok: 15.00},
	"claude-haiku-4-5":  {InputPerMTok: 1.00, OutputPerMTok: 5.00},

	// OpenAI (approximate)
	"gpt-5.1":      {InputPerMTok: 1.25, OutputPerMTok: 10.00},
	"gpt-5":        {InputPerMTok: 1.25, OutputPerMTok: 10.00},
	"gpt-5-mini":   {InputPerMTok: 0.25, OutputPerMTok: 2.00},
	"gpt-5-nano":   {InputPerMTok: 0.05, OutputPerMTok: 0.40},
	"gpt-4.1":      {InputPerMTok: 2.00, OutputPerMTok: 8.00},
	"gpt-4.1-mini": {InputPerMTok: 0.40, OutputPerMTok: 1.60},

	// Google (approximate)
	"gemini-3-pro":     {InputPerMTok: 2.00, OutputPerMTok: 12.00},
	"gemini-2.5-pro":   {InputPerMTok: 1.25, OutputPerMTok: 10.00},
	"gemini-2.5-flash": {InputPerMTok: 0.30, OutputPerMTok: 2.50},
}

// fallbackPricing is used for unknown models: assume a modest offering.
var fallbackPricing = Pricing{InputPerMTok: 1.00, OutputPerMTok: 5.00}

func (c Config) pricingFor(model string) Pricing {
	if p, ok := c.PricingOverride[model]; ok {
		return p
	}
	if p, ok := builtinPricing[model]; ok {
		return p
	}
	return fallbackPricing
}

// sacrificeUSD is the market value of the tokens offered up.
func (p Pricing) sacrificeUSD(inputTokens, outputTokens int64) float64 {
	return float64(inputTokens)*p.InputPerMTok/1e6 +
		float64(outputTokens)*p.OutputPerMTok/1e6
}
