package seed

import "gorm.io/gorm"

// Run executes all seed functions.
func Run(db *gorm.DB) {
	seedUsers(db)
	seedQuizzes(db)
}
