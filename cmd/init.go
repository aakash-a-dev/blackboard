package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Write a starter api.yaml in the current directory",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

const starterTemplate = `# blackboard — starter template
# Run with: blackboard serve api.yaml

version: "1"
info:
  title: My Mock API
  base_path: /api/v1

endpoints:
  # List users
  - path: /users
    method: GET
    response:
      status: 200
      body:
        - id: "{{uuid}}"
          name: "{{fullName}}"
          email: "{{email}}"
          role: "user"
          created_at: "{{isoDate}}"
      count: 8

  # Get single user by ID
  - path: /users/:id
    method: GET
    response:
      status: 200
      body:
        id: "{{param.id}}"
        name: "{{fullName}}"
        email: "{{email}}"
        role: "user"

  # Create a user
  - path: /users
    method: POST
    request:
      required_fields: [name, email]
      body_schema:
        name:
          type: string
          min_length: 2
        email:
          type: string
          format: email
    response:
      status: 201
      body:
        id: "{{uuid}}"
        message: "User created successfully"

  # Delete a user
  - path: /users/:id
    method: DELETE
    response:
      status: 204

  # Simulate an unreliable endpoint
  - path: /integrations/webhook
    method: POST
    behavior:
      error_rate: 0.2
      error_status: 502
      error_body:
        message: "Bad gateway"
    response:
      status: 200
      body:
        received: true

  # Paginated list
  - path: /posts
    method: GET
    response:
      status: 200
      paginated: true
      body:
        - id: "{{uuid}}"
          title: "{{sentence}}"
          author: "{{fullName}}"
      total: 50
`

func runInit(cmd *cobra.Command, args []string) error {
	const target = "api.yaml"

	// Never clobber — decision #20
	if _, err := os.Stat(target); err == nil {
		return errors.New(`api.yaml already exists. Delete it first or rename it`)
	}

	if err := os.WriteFile(target, []byte(starterTemplate), 0644); err != nil {
		return err
	}

	cmd.Printf("  Created %s\n  Run: blackboard serve %s\n", target, target)
	return nil
}
