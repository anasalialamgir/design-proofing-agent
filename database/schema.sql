CREATE TYPE submission_status AS ENUM (
    'PENDING_INGESTION',
    'INGESTED',
    'PROCESSING_PREFLIGHT',
    'PREFLIGHT_PASSED',
    'PREFLIGHT_WARNING',
    'PREFLIGHT_FAILED',
    'CORRUPT_FILE',
    'PROOF_GENERATED',
    'AWAITING_APPROVAL',
    'APPROVED',
    'REJECTED',
    'REVISION_REQUESTED'
);

CREATE TYPE issue_severity AS ENUM ('INFO', 'WARNING', 'CRITICAL');
CREATE TYPE color_space_type AS ENUM ('CMYK', 'RGB', 'GRAYSCALE', 'PANTONE_SPOT', 'MIXED');

CREATE TABLE artwork_submissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id VARCHAR(64) NOT NULL,
    customer_id VARCHAR(64) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_sha256 CHAR(64) NOT NULL,
    gcs_raw_uri TEXT NOT NULL,
    status submission_status NOT NULL DEFAULT 'PENDING_INGESTION',
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE preflight_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES artwork_submissions(id) ON DELETE CASCADE,
    detected_format VARCHAR(32) NOT NULL,
    is_vector BOOLEAN NOT NULL DEFAULT FALSE,
    page_count INT NOT NULL DEFAULT 1,
    width_pts NUMERIC(10, 2) NOT NULL,
    height_pts NUMERIC(10, 2) NOT NULL,
    color_space color_space_type NOT NULL,
    min_effective_dpi INT,
    is_corrupt BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE preflight_defects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID NOT NULL REFERENCES preflight_reports(id) ON DELETE CASCADE,
    severity issue_severity NOT NULL,
    defect_code VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    page_number INT NOT NULL DEFAULT 1
);

CREATE TABLE proof_artifacts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id UUID NOT NULL REFERENCES artwork_submissions(id) ON DELETE CASCADE,
    gcs_proof_pdf_uri TEXT NOT NULL,
    gcs_preview_png_uri TEXT NOT NULL,
    approval_token VARCHAR(128) NOT NULL UNIQUE,
    decision_notes TEXT,
    approved_by_email VARCHAR(255),
    decided_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(64) NOT NULL,
    payload JSONB NOT NULL,
    dispatched BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
