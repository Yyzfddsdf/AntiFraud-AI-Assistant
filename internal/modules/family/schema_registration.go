package family_system

import "antifraud/internal/platform/database"

func init() {
	database.RegisterMainDBSchemaInitializer("family_system", EnsureSchema)
}
