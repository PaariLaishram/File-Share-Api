package models

import "FileShare/database"

type Profile struct {
	User  *int    `json:"user,omitempty"`
	Email *string `json:"email,omitempty"`
}

func (profile Profile) GetProfile(user string) ([]any, error) {
	get_profile_query := `SELECT id, email FROM users where user = ?`
	profile_fields := []any{&profile.User, &profile.Email}
	profile_filter := []any{user}
	return database.RunQuery(get_profile_query, profile_fields, profile_filter, &profile)
}
