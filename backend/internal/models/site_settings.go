package models

import "time"

// SiteSetting menyimpan pengaturan situs dalam format key-value
type SiteSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SiteSettingsResponse mengembalikan semua pengaturan situs
type SiteSettingsResponse struct {
	AppName    string `json:"app_name"`
	AppDesc    string `json:"app_description"`
	AboutText  string `json:"about_text"`
	LogoURL    string `json:"logo_url"`
	Address    string `json:"address"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	FooterText string `json:"footer_text"`
}
