-- Create fakultas table
CREATE TABLE IF NOT EXISTS fakultas (
    id UUID PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    kode VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create prodi table
CREATE TABLE IF NOT EXISTS prodi (
    id UUID PRIMARY KEY,
    nama VARCHAR(255) NOT NULL,
    kode VARCHAR(50) NOT NULL UNIQUE,
    fakultas_id UUID NOT NULL REFERENCES fakultas(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_fakultas_kode ON fakultas(kode);
CREATE INDEX IF NOT EXISTS idx_prodi_kode ON prodi(kode);
CREATE INDEX IF NOT EXISTS idx_prodi_fakultas_id ON prodi(fakultas_id);

-- Update jenis_file constraint on documents (if any)
-- Ensure jenis_file column supports: skripsi, tesis, jurnal
