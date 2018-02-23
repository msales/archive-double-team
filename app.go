package double_team

// Application represents the application.
type Application struct {}

// NewApplication creates an instance of Application.
func NewApplication() *Application {
	return &Application{}
}

// IsHealthy checks the health of the Application.
func (a *Application) IsHealthy() error {
	return nil
}
