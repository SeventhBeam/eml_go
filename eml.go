package eml

type EML struct {
	*Dependencies
}

type Dependencies struct {
	Config *Config
	Store  Store
	//Data      data.Store
	//Users     users.Store
	//Secrets   secrets.Store
	//Messaging messages.Service
	//Queueing  queues.Service
	//Validator validation.Validator
	//Env       *env.Settings
}

func New(d *Dependencies) *EML {
	return &EML{
		d,
	}
}