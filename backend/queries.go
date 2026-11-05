package main

const (
	qGetEmployees = "SELECT id, name, email, department FROM employees ORDER BY id"

	qInsertEmployees = `INSERT INTO employees (name, email, department) VALUES ($1, $2, $3)
			            RETURNING id, name, email, department`

	qDeleteEmployeesById = `DELETE FROM employees WHERE id = $1`
)
