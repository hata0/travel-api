package response

type (
	Trip struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	GetTripResponse struct {
		Trip Trip `json:"trip"`
	}

	ListTripResponse struct {
		Trips []Trip `json:"trips"`
	}
)
