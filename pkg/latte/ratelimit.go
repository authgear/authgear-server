package latte

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/interaction"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
)

var AntiSpamRequestBucket = interaction.AntiSpamRequestBucket
var AntiSpamSignupBucket = interaction.AntiSpamSignupBucket
var AntiSpamSignupAnonymousBucket = interaction.AntiSpamSignupAnonymousBucket
var AntiAccountEnumerationBucket = interaction.AntiAccountEnumerationBucket

func AntiSpamEmailOTPCodeBucket(emailConfig *config.EmailConfig, target string) ratelimit.Bucket {
	maker := interaction.AntiSpamOTPCodeBucketMaker{EmailConfig: emailConfig}
	return maker.MakeBucket(model.AuthenticatorOOBChannelEmail, target)
}

func AntiSpamSMSOTPCodeBucket(smsConfig *config.SMSConfig, target string) ratelimit.Bucket {
	maker := interaction.AntiSpamOTPCodeBucketMaker{SMSConfig: smsConfig}
	return maker.MakeBucket(model.AuthenticatorOOBChannelSMS, target)
}
