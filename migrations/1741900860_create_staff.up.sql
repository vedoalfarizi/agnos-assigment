-- Create staff table
CREATE TABLE staff (
  id SERIAL PRIMARY KEY,
  hospital_id INTEGER NOT NULL REFERENCES hospital(id),
  username VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Create index on hospital_id for efficient filtering
CREATE INDEX idx_staff_hospital_id ON staff(hospital_id);
