-- Create sequence for apps
CREATE TABLE apps (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    base_url VARCHAR(255) NOT NULL,
    connection_string TEXT,
    database_name VARCHAR(255),
    maintenance_mode BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_apps_name UNIQUE (name)
);

CREATE TABLE job_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_job_status_name UNIQUE (name)
);

CREATE TABLE task_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_task_status_name UNIQUE (name)
);

CREATE TABLE channel_task_status (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_channel_task_status_name UNIQUE (name)
);

CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_channels_name UNIQUE (name)
);

CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL REFERENCES apps(id) ON DELETE RESTRICT,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status_id INT NOT NULL REFERENCES job_status(id) ON DELETE RESTRICT,
    max_thread_count INT NOT NULL DEFAULT 1,
    arguments JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_jobs_app_id ON jobs(app_id);
CREATE INDEX idx_jobs_status_id ON jobs(status_id);
CREATE INDEX idx_jobs_app_id_name ON jobs(app_id, name);

CREATE TABLE job_channels (
    id SERIAL PRIMARY KEY,
    job_id INT NOT NULL REFERENCES jobs(id) ON DELETE RESTRICT,
    channel_id INT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_job_channels_job_channel UNIQUE (job_id, channel_id)
);
CREATE INDEX idx_job_channels_job_id ON job_channels(job_id);
CREATE INDEX idx_job_channels_channel_id ON job_channels(channel_id);

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    job_id INT NOT NULL REFERENCES jobs(id) ON DELETE RESTRICT,
    parent_task_id INT REFERENCES tasks(id) ON DELETE RESTRICT,
    status_id INT NOT NULL REFERENCES task_status(id) ON DELETE RESTRICT,
    arguments JSONB,
    task_trigger_time TIMESTAMP NOT NULL,
    task_start_time TIMESTAMP,
    task_end_time TIMESTAMP,
    current_retry_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_tasks_job_id ON tasks(job_id);
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id);
CREATE INDEX idx_tasks_status_id ON tasks(status_id);
CREATE INDEX idx_tasks_task_trigger_time ON tasks(task_trigger_time);
CREATE INDEX idx_tasks_status_id_trigger_time ON tasks(status_id, task_trigger_time);

CREATE TABLE channel_tasks (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE RESTRICT,
    channel_id INT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    status_id INT NOT NULL REFERENCES channel_task_status(id) ON DELETE RESTRICT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uk_channel_tasks_task_channel UNIQUE (task_id, channel_id)
);
CREATE INDEX idx_channel_tasks_task_id ON channel_tasks(task_id);
CREATE INDEX idx_channel_tasks_channel_id ON channel_tasks(channel_id);
CREATE INDEX idx_channel_tasks_status_id ON channel_tasks(status_id);

CREATE TABLE task_logs (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE RESTRICT,
    job_id INT NOT NULL REFERENCES jobs(id) ON DELETE RESTRICT,
    step_name VARCHAR(255) NOT NULL,
    performance_log JSONB,
    error_log JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_task_logs_task_id ON task_logs(task_id);
CREATE INDEX idx_task_logs_job_id ON task_logs(job_id);
CREATE INDEX idx_task_logs_created_at ON task_logs(created_at);

-- Seed Data

INSERT INTO job_status (name) VALUES 
('Active'), 
('Disabled'), 
('Archived');

INSERT INTO task_status (name) VALUES 
('NeedToPick'), 
('Picked'), 
('Processing'), 
('Completed'), 
('Failed'), 
('RetryScheduled'), 
('Cancelled');

INSERT INTO channel_task_status (name) VALUES 
('Pending'), 
('Sent'), 
('Failed'), 
('RetryScheduled');
