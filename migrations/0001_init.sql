CREATE TABLE repositories (
    id TEXT PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL,
    remote_url TEXT,
    default_branch TEXT,
    head_commit_sha TEXT,
    config_version INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE index_runs (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    head_commit_sha TEXT,
    files_scanned INTEGER NOT NULL DEFAULT 0,
    files_indexed INTEGER NOT NULL DEFAULT 0,
    files_failed INTEGER NOT NULL DEFAULT 0,
    symbols_found INTEGER NOT NULL DEFAULT 0,
    relations_found INTEGER NOT NULL DEFAULT 0,
    warnings TEXT,
    error_message TEXT
);

CREATE TABLE source_files (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    path TEXT NOT NULL,
    language TEXT,
    kind TEXT NOT NULL,
    content_hash TEXT NOT NULL,
    size_bytes INTEGER NOT NULL,
    line_count INTEGER,
    is_generated INTEGER NOT NULL DEFAULT 0,
    is_test_file INTEGER NOT NULL DEFAULT 0,
    is_ignored INTEGER NOT NULL DEFAULT 0,
    last_seen_index_run_id TEXT NOT NULL REFERENCES index_runs(id),
    deleted_at TEXT,
    UNIQUE (repository_id, path)
);

CREATE TABLE code_symbols (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    source_file_id TEXT NOT NULL REFERENCES source_files(id),
    parent_symbol_id TEXT REFERENCES code_symbols(id),
    component_id TEXT,
    qualified_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    kind TEXT NOT NULL,
    visibility TEXT,
    signature TEXT,
    start_line INTEGER NOT NULL,
    start_column INTEGER,
    end_line INTEGER NOT NULL,
    end_column INTEGER,
    semantic_hash TEXT NOT NULL,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    last_seen_index_run_id TEXT NOT NULL REFERENCES index_runs(id),
    deleted_at TEXT
);

CREATE TABLE code_relations (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    from_symbol_id TEXT NOT NULL REFERENCES code_symbols(id),
    to_symbol_id TEXT REFERENCES code_symbols(id),
    to_external_ref TEXT,
    kind TEXT NOT NULL,
    source_file_id TEXT REFERENCES source_files(id),
    line INTEGER,
    metadata TEXT,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    last_seen_index_run_id TEXT NOT NULL REFERENCES index_runs(id),
    deleted_at TEXT
);

CREATE TABLE entry_points (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    handler_symbol_id TEXT REFERENCES code_symbols(id),
    component_id TEXT,
    kind TEXT NOT NULL,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    method TEXT,
    path TEXT,
    topic TEXT,
    schedule TEXT,
    command TEXT,
    framework TEXT,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    last_seen_index_run_id TEXT NOT NULL REFERENCES index_runs(id),
    deleted_at TEXT,
    UNIQUE (repository_id, key)
);

CREATE TABLE stories (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    key TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT,
    intent TEXT NOT NULL,
    outcome TEXT,
    status TEXT NOT NULL,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    verification_status TEXT NOT NULL,
    last_verified_at TEXT,
    last_verified_index_run_id TEXT REFERENCES index_runs(id),
    owner TEXT,
    tags TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    archived_at TEXT,
    UNIQUE (repository_id, key)
);

CREATE TABLE story_triggers (
    id TEXT PRIMARY KEY NOT NULL,
    story_id TEXT NOT NULL REFERENCES stories(id),
    entry_point_id TEXT REFERENCES entry_points(id),
    kind TEXT NOT NULL,
    label TEXT NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL
);

CREATE TABLE story_actors (
    id TEXT PRIMARY KEY NOT NULL,
    story_id TEXT NOT NULL REFERENCES stories(id),
    component_id TEXT,
    external_ref TEXT,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    kind TEXT NOT NULL,
    description TEXT,
    visual_style TEXT,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    sort_order INTEGER NOT NULL,
    UNIQUE (story_id, key)
);

CREATE TABLE scenes (
    id TEXT PRIMARY KEY NOT NULL,
    story_id TEXT NOT NULL REFERENCES stories(id),
    key TEXT NOT NULL,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    narration TEXT,
    technical_summary TEXT,
    from_actor_id TEXT REFERENCES story_actors(id),
    to_actor_id TEXT REFERENCES story_actors(id),
    primary_symbol_id TEXT REFERENCES code_symbols(id),
    primary_component_id TEXT,
    operation TEXT,
    condition TEXT,
    input_summary TEXT,
    output_summary TEXT,
    status TEXT NOT NULL,
    confidence TEXT NOT NULL,
    visual_metadata TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (story_id, key)
);

CREATE TABLE story_paths (
    id TEXT PRIMARY KEY NOT NULL,
    story_id TEXT NOT NULL REFERENCES stories(id),
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    kind TEXT NOT NULL,
    description TEXT,
    entry_scene_id TEXT REFERENCES scenes(id),
    exit_scene_id TEXT REFERENCES scenes(id),
    is_default INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL,
    UNIQUE (story_id, key)
);

CREATE TABLE scene_transitions (
    id TEXT PRIMARY KEY NOT NULL,
    story_id TEXT NOT NULL REFERENCES stories(id),
    from_scene_id TEXT NOT NULL REFERENCES scenes(id),
    to_scene_id TEXT NOT NULL REFERENCES scenes(id),
    kind TEXT NOT NULL,
    label TEXT,
    condition TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    is_default INTEGER NOT NULL DEFAULT 0,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL
);

CREATE TABLE evidences (
    id TEXT PRIMARY KEY NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id),
    kind TEXT NOT NULL,
    source_file_id TEXT REFERENCES source_files(id),
    code_symbol_id TEXT REFERENCES code_symbols(id),
    test_case_id TEXT,
    contract_id TEXT,
    document_id TEXT,
    git_commit_id TEXT,
    entry_point_id TEXT REFERENCES entry_points(id),
    external_ref TEXT,
    locator TEXT NOT NULL,
    content_hash TEXT,
    snapshot_hash TEXT,
    title TEXT,
    excerpt TEXT,
    source_type TEXT NOT NULL,
    confidence TEXT NOT NULL,
    last_seen_index_run_id TEXT REFERENCES index_runs(id),
    deleted_at TEXT
);

CREATE TABLE evidence_references (
    id TEXT PRIMARY KEY NOT NULL,
    evidence_id TEXT NOT NULL REFERENCES evidences(id),
    story_id TEXT REFERENCES stories(id),
    scene_id TEXT REFERENCES scenes(id),
    invariant_id TEXT,
    failure_mode_id TEXT,
    role TEXT NOT NULL,
    claim TEXT,
    is_primary INTEGER NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL,
    added_by TEXT NOT NULL,
    created_at TEXT NOT NULL
);

CREATE TABLE evidence_verifications (
    id TEXT PRIMARY KEY NOT NULL,
    evidence_id TEXT NOT NULL REFERENCES evidences(id),
    index_run_id TEXT NOT NULL REFERENCES index_runs(id),
    status TEXT NOT NULL,
    expected_hash TEXT,
    actual_hash TEXT,
    message TEXT,
    verified_at TEXT NOT NULL
);

CREATE INDEX idx_source_files_repository_path ON source_files (repository_id, path);
CREATE INDEX idx_code_symbols_repository_qualified ON code_symbols (repository_id, qualified_name);
CREATE INDEX idx_code_symbols_file_start ON code_symbols (source_file_id, start_line);
CREATE INDEX idx_code_relations_from_kind ON code_relations (from_symbol_id, kind);
CREATE INDEX idx_code_relations_to_kind ON code_relations (to_symbol_id, kind);
CREATE INDEX idx_entry_points_repository_key ON entry_points (repository_id, key);
CREATE INDEX idx_stories_repository_key ON stories (repository_id, key);
CREATE INDEX idx_stories_repository_status ON stories (repository_id, status);
CREATE INDEX idx_scenes_story_key ON scenes (story_id, key);
CREATE INDEX idx_scene_transitions_from ON scene_transitions (from_scene_id);
CREATE INDEX idx_evidences_repository_kind ON evidences (repository_id, kind);
CREATE INDEX idx_evidence_references_scene ON evidence_references (scene_id);
CREATE INDEX idx_evidence_references_story ON evidence_references (story_id);
CREATE INDEX idx_evidence_verifications_evidence_at ON evidence_verifications (evidence_id, verified_at);
