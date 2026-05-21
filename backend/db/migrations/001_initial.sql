-- Teachers
CREATE TABLE IF NOT EXISTS teachers (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Teacher scheduling settings
CREATE TABLE IF NOT EXISTS teacher_settings (
    teacher_id              UUID PRIMARY KEY REFERENCES teachers(id) ON DELETE CASCADE,
    working_start           TIME NOT NULL DEFAULT '08:00',
    working_end             TIME NOT NULL DEFAULT '20:00',
    lesson_duration_minutes INT  NOT NULL DEFAULT 45,
    max_gap_minutes         INT  NOT NULL DEFAULT 30,
    max_consecutive_lessons INT  NOT NULL DEFAULT 4,
    break_after_n_lessons   INT  NOT NULL DEFAULT 0,   -- 0 = disabled
    break_duration_minutes  INT  NOT NULL DEFAULT 15,
    locale                  TEXT NOT NULL DEFAULT 'en'
);

-- Students
CREATE TABLE IF NOT EXISTS students (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    email      TEXT,
    notes      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Availability slots (shared for teacher and student)
CREATE TABLE IF NOT EXISTS availability_slots (
    id         UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type TEXT    NOT NULL CHECK (owner_type IN ('teacher', 'student')),
    owner_id   UUID    NOT NULL,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Mon, 6=Sun
    start_time TIME    NOT NULL,
    end_time   TIME    NOT NULL,
    CHECK (start_time < end_time)
);

CREATE INDEX IF NOT EXISTS idx_availability_owner
    ON availability_slots(owner_type, owner_id);

-- Weekly timetable lessons
CREATE TABLE IF NOT EXISTS lessons (
    id         UUID     PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID     NOT NULL REFERENCES teachers(id) ON DELETE CASCADE,
    student_id UUID     NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time TIME     NOT NULL,
    end_time   TIME     NOT NULL,
    is_pinned  BOOLEAN  NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_time < end_time)
);

CREATE INDEX IF NOT EXISTS idx_lessons_teacher ON lessons(teacher_id);
CREATE INDEX IF NOT EXISTS idx_lessons_student ON lessons(student_id);
