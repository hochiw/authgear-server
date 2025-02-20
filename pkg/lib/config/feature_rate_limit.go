package config

var _ = FeatureConfigSchema.Add("RateLimitBucketConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"size": { "type": "integer", "minimum": 0 },
		"reset_period": { "type": "integer", "minimum": 0 }
	}
}
`)

var _ = FeatureConfigSchema.Add("RateLimitFeatureConfig", `
{
	"type": "object",
	"additionalProperties": false,
	"properties": {
		"sms": { "$ref": "#/$defs/RateLimitBucketConfig" }
	}
}
`)

type RateLimitFeatureConfig struct {
	SMS *SMSRateLimitBucketConfig `json:"sms,omitempty"`
}

type SMSRateLimitBucketConfig struct {
	Size        *int             `json:"size,omitempty"`
	ResetPeriod *DurationSeconds `json:"reset_period,omitempty"`
}

func (c *SMSRateLimitBucketConfig) SetDefaults() {
	if c.Size == nil {
		c.Size = newInt(100000)
	}
	if c.ResetPeriod == nil {
		a := DurationSeconds(30 * 24 * 60 * 60)
		c.ResetPeriod = &a
	}
}
