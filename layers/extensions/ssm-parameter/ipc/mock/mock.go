package mock

//go:generate mockgen -destination ./ssmiface.go -package mock github.com/aws/aws-sdk-go/service/ssm/ssmiface SSMAPI
