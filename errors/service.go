package errors

// Service is a identifier which service throw the error
// mainly used to determine whether internal application / external
type Service string

// General service which SkorLife app using
// if any other service we are using; with Service type we can extend it
// It's recommended to add in the package level
const (
	SKORLIFE    Service = "SKORLIFE"
	VIDA        Service = "VIDA"
	CLIK        Service = "CLIK"
	AKSATA      Service = "AKSATA"
	AWS         Service = "AWS"
	INFOBIP     Service = "INFOBIP"
	JIRA        Service = "JIRA"
	RABBIT_MQ   Service = "RABBIT_MQ"
	REDIS       Service = "REDIS"
	POSTGRES    Service = "POSTGRES"
	MYSQL       Service = "MYSQL"
	MONGO       Service = "MONGO"
	FIREBASE    Service = "FIREBASE"
	ADVANCED_AI Service = "ADVANCED_AI"
)
