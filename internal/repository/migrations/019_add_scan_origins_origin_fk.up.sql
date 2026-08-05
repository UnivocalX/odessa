ALTER TABLE scan_origins
ADD CONSTRAINT fk_scan_origins_origin
FOREIGN KEY (origin_id) REFERENCES origins(id) ON DELETE CASCADE NOT VALID;
