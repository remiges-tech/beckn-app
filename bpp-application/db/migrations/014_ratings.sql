-- Ratings submitted by BAPs for any ratable entity:
-- contract, fulfillment/performance, item/resource, provider, agent, support interaction.
-- target_type/target_id identify the rated entity; range captures the score.
-- feedback_form_submission stores the filled feedback form (FormSubmission schema).

CREATE TABLE ratings (
    id                          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Context linkage
    contract_id                 UUID        NOT NULL REFERENCES contracts (id),
    bap_id                      TEXT        NOT NULL,
    bap_uri                     TEXT        NOT NULL,

    -- The entity being rated
    target_id                   TEXT        NOT NULL,
    target_descriptor           JSONB,
    target_attributes           JSONB,

    -- Score range (domain-specific; stored as JSONB for flexibility)
    -- e.g. {min: 1, max: 5, value: 4.5} or {thumbs: "up"}
    range                       JSONB       NOT NULL,

    -- Filled feedback form (FormSubmission: id, submissionId, data, submittedAt)
    feedback_form_submission    JSONB,

    -- Whether this was a preview (context.try: true) or committed rating
    is_preview                  BOOLEAN     NOT NULL DEFAULT FALSE,

    created_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE ratings IS
    'Rating inputs from BAPs for any ratable entity within a completed contract.';

COMMENT ON COLUMN ratings.range IS
    'Score expressed as JSONB (min, max, value) to support any rating scale.';

COMMENT ON COLUMN ratings.feedback_form_submission IS
    'Serialized FormSubmission including form URI, submission ID, and field responses.';

---- create above / drop below ----

DROP TABLE IF EXISTS ratings;
