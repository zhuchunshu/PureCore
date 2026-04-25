package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var makeModelCmd = &cobra.Command{
	Use:   "make:model [name]",
	Short: "Create a new model file",
	Long:  `Create a new GORM model file in app/Models/.`,
	Args:  cobra.ExactArgs(1),
	Run:   makeModelRun,
}

var makeControllerCmd = &cobra.Command{
	Use:   "make:controller [name]",
	Short: "Create a new controller file",
	Long:  `Create a new controller file in app/Http/Controllers/ with CRUD scaffold.`,
	Args:  cobra.ExactArgs(1),
	Run:   makeControllerRun,
}

var makeMigrationCmd = &cobra.Command{
	Use:   "make:migration [name]",
	Short: "Create a new migration file",
	Long:  `Create a new migration file in database/migrations/.`,
	Args:  cobra.ExactArgs(1),
	Run:   makeMigrationRun,
}

func init() {
	rootCmd.AddCommand(makeModelCmd)
	rootCmd.AddCommand(makeControllerCmd)
	rootCmd.AddCommand(makeMigrationCmd)
}

func makeModelRun(cmd *cobra.Command, args []string) {
	name := args[0]
	fileName := name + ".go"
	filePath := filepath.Join("app", "Models", fileName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filePath), 0755)

	content := fmt.Sprintf(`package models

import "purecore/core"

// %s represents a %s record in the database
type %s struct {
	core.Model
	// Add your fields here
	Name string `+"`"+`gorm:"type:varchar(100);not null" json:"name"`+"`"+`
}
`, name, strings.ToLower(name), name)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("✗ Failed to create model: %v\n", err)
		return
	}
	fmt.Printf("✓ Model created: %s\n", filePath)
}

func makeControllerRun(cmd *cobra.Command, args []string) {
	name := args[0]
	fileName := name + "Controller.go"
	filePath := filepath.Join("app", "Http", "Controllers", fileName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filePath), 0755)

	content := fmt.Sprintf(`package controllers

import (
	models "purecore/app/Models"
	"purecore/core"
)

type %sController struct{}

type Create%sRequest struct {
	Name string `+"`"+`json:"name" validate:"required,min=2"`+"`"+`
}

func (c *%sController) Index(req *core.Request, res *core.Response) error {
	var records []models.%s
	if err := core.DB().Find(&records).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(records)
}

func (c *%sController) Store(req *core.Request, res *core.Response) error {
	var body Create%sRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error())
	}

	record := models.%s{Name: body.Name}
	if err := core.DB().Create(&record).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(record)
}

func (c *%sController) Show(req *core.Request, res *core.Response) error {
	id := req.Input("id")
	if id == "" {
		return res.NotFound("Not found")
	}

	var record models.%s
	if err := core.DB().First(&record, id).Error; err != nil {
		return res.NotFound("Not found")
	}
	return res.Success(record)
}
`, name, name, name, name, name, name, name, name, name)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("✗ Failed to create controller: %v\n", err)
		return
	}
	fmt.Printf("✓ Controller created: %s\n", filePath)
}

func makeMigrationRun(cmd *cobra.Command, args []string) {
	name := args[0]

	// Generate a timestamped migration file
	timestamp := time.Now().Format("2006_01_02_150405")
	fileName := fmt.Sprintf("%s_create_%s_table.go", timestamp, strings.ToLower(name))
	filePath := filepath.Join("database", "migrations", fileName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filePath), 0755)

	content := fmt.Sprintf(`package migrations

import (
	"purecore/core"

	"gorm.io/gorm"
)

func init() {
	core.RegisterMigration("%s", up%s)
}

func up%s(db *gorm.DB) error {
	type %s struct {
		core.Model
		Name string `+"`"+`gorm:"type:varchar(100);not null"`+"`"+`
	}
	return db.AutoMigrate(&%s{})
}
	`, fileName, name, name, name, name)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("✗ Failed to create migration: %v\n", err)
		return
	}
	fmt.Printf("✓ Migration created: %s\n", filePath)
}
