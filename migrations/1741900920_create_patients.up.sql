-- Create patients table
CREATE TABLE patients (
  id SERIAL PRIMARY KEY,
  national_id VARCHAR(50) UNIQUE,
  passport_id VARCHAR(50) UNIQUE,
  first_name_th VARCHAR(255),
  middle_name_th VARCHAR(255),
  last_name_th VARCHAR(255),
  first_name_en VARCHAR(255),
  middle_name_en VARCHAR(255),
  last_name_en VARCHAR(255),
  date_of_birth DATE,
  phone_number VARCHAR(20) UNIQUE,
  email VARCHAR(255) UNIQUE,
  gender VARCHAR(1),
  hospital_id INTEGER NOT NULL REFERENCES hospital(id),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  CHECK (national_id IS NOT NULL OR passport_id IS NOT NULL)
);

-- Create indexes for data isolation and search performance
CREATE INDEX idx_patients_hospital_id ON patients(hospital_id);
CREATE INDEX idx_patients_dob ON patients(date_of_birth);
