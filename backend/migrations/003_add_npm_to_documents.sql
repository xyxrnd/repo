-- Add NPM (Nomor Pokok Mahasiswa) column to documents table
ALTER TABLE documents ADD COLUMN IF NOT EXISTS npm VARCHAR(20) DEFAULT '';
