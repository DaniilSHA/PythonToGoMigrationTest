package main

import "PythonToGoMigrationTest/internal/generator"

// --url http://127.0.0.1:8080/calc --threads 10 --interval 0.1 --timeout 5
func main() {
	generator.Start()
}
