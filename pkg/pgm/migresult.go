package pgm

type MigrationResultStatus string

const (
	APPLIED  MigrationResultStatus = "applied"
	REVERTED MigrationResultStatus = "reverted"
)

type MigrationResult struct {
	MigrationName string                `json:"migrationName"`
	Status        MigrationResultStatus `json:"status"`
}
