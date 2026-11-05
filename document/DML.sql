INSERT INTO employees (name, email, department)
VALUES
    ('Tony Stark', 'tony@avengers.com', 'IT'),
    ('Steve Rogers', 'steve@avengers.com', 'HR')
ON CONFLICT (email) DO NOTHING;
