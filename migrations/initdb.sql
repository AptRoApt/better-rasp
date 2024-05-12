CREATE TABLE IF NOT EXISTS faculties (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS education_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    faculty_id INTEGER REFERENCES faculties(id),
    course INTEGER NOT NULL,
    education_type_id INTEGER REFERENCES education_types
);

CREATE TABLE IF NOT EXISTS cathedras (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS teachers (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS m2m_teachers_cathedras (
    teacher_id INTEGER REFERENCES teachers,
    cathedra_id INTEGER REFERENCES cathedras,
    PRIMARY KEY(teacher_id, cathedra_id)
);

CREATE TABLE IF NOT EXISTS rooms (
    id SERIAL PRIMARY KEY,
    building_num INTEGER NOT NULL,
    num text NOT NULL,
    UNIQUE (building_num, num)
);

CREATE TABLE IF NOT EXISTS lesson_time (
    num INTEGER PRIMARY KEY,
    start_time TIME,
    end_time TIME
);

INSERT INTO lesson_time (num, start_time, end_time)
VALUES 
    (1, '08:30:00', '10:00:00'),
    (2, '10:10:00', '11:40:00'),
    (3, '11:50:00', '13:20:00'),
    (4, '14:00:00', '15:30:00'),
    (5, '15:40:00', '17:10:00'),
    (6, '17:20:00', '18:50:00'),
    (7, '18:55:00', '20:25:00'),
    (8, '20:30:00', '22:00:00')
    ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS lesson_types (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS disciplines (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS lessons (
    id SERIAL PRIMARY KEY,
    reaid BIGINT,
    date DATE NOT NULL,
    lesson_num INTEGER REFERENCES lesson_time,
    lesson_type_id INTEGER REFERENCES lesson_types,
    discipline_id INTEGER REFERENCES disciplines,
    room_id INTEGER REFERENCES rooms,
    subgroup_num INTEGER NOT NULL,
    cathedra_id INTEGER REFERENCES cathedras,
    is_commission BOOLEAN DEFAULT FALSE,
    UNIQUE (reaid, room_id)
);

CREATE TABLE IF NOT EXISTS m2m_groups_lessons (
    group_id INTEGER REFERENCES groups,
    lesson_id INTEGER REFERENCES lessons,
    PRIMARY KEY(group_id, lesson_id)
);

CREATE TABLE IF NOT EXISTS m2m_teachers_lessons (
    teacher_id INTEGER REFERENCES teachers,
    lesson_id INTEGER REFERENCES lessons,
    PRIMARY KEY (teacher_id, lesson_id)
);