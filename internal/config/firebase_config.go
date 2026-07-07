package config

// FirebaseConfig holds settings required to authenticate with Firebase services
// (e.g. Firebase Cloud Messaging for push notifications).
//
// All fields are optional during Phase 2 startup. Validation is deferred to the
// Firebase provider initialization phase when the provider is actually used.
type FirebaseConfig struct {
	// CredentialsFile is the path to the Firebase service-account JSON key file.
	// Example: /etc/secrets/firebase-credentials.json
	CredentialsFile string

	// ProjectID is the Google Cloud / Firebase project identifier.
	ProjectID string
}

// loadFirebaseConfig reads Firebase settings from environment variables.
// Both fields are optional; missing values result in an empty config.
func loadFirebaseConfig() (FirebaseConfig, error) {
	return FirebaseConfig{
		CredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", ""),
		ProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
	}, nil
}
