package family_system

import "antifraud/database"

func init() {
	database.RegisterMainDBSchemaInitializer("family_system", EnsureSchema)
}
