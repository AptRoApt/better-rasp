-- Создание таблицы факультетов
CREATE TABLE IF NOT EXISTS faculties (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

--Создание таблицы типов обучения
CREATE TABLE IF NOT EXISTS education_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Создание таблицы groups
CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    faculty_id INT REFERENCES faculties(id),
    course INT NOT NULL,
    education_type_id INT REFERENCES education_types(id)
);

CREATE TABLE IF NOT EXISTS departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

-- Создание таблицы teachers
CREATE TABLE IF NOT EXISTS teachers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    department_id INT REFERENCES departments(id)
);

-- Создание таблицы buildings
CREATE TABLE IF NOT EXISTS buildings (
    num INT PRIMARY KEY,
    address VARCHAR(255)
);

-- Создание таблицы rooms
CREATE TABLE IF NOT EXISTS rooms (
    building_num INT REFERENCES buildings(num) ON DELETE CASCADE,
    num INT NOT NULL,
    PRIMARY KEY(building_num, num)
);

-- Создание таблицы lesson_time
CREATE TABLE IF NOT EXISTS lesson_time (
    id INT PRIMARY KEY,
    start_time TIME,
    end_time TIME
);

-- Вставка значений в таблицу lesson_time
INSERT INTO lesson_time (id, start_time, end_time)
VALUES 
    (1, '08:30:00', '10:00:00'),
    (2, '10:10:00', '11:40:00'),
    (3, '11:50:00', '13:20:00'),
    (4, '14:00:00', '15:30:00'),
    (5, '15:40:00', '17:10:00'),
    (6, '17:20:00', '18:50:00'),
    (7, '18:55:00', '20:25:00'),
    (8, '20:30:00', '22:00:00');

-- Создание таблицы lesson_types
CREATE TABLE IF NOT EXISTS lesson_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS disciplines (
    id SERIAL PRIMARY KEY,
    name VARCHAR (255) NOT NULL
);

-- Создание таблицы lessons
CREATE TABLE IF NOT EXISTS lessons (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    lesson_num INT REFERENCES lesson_time(id),
    lesson_type_id INT REFERENCES lesson_types(id),
    discipline_id INT REFERENCES disciplines(id),
    teacher_id INT REFERENCES teachers(id) ON DELETE SET NULL,
    subgroup_num INT
);

-- Создание таблицы m2m_groups_lessons
CREATE TABLE IF NOT EXISTS m2m_groups_lessons (
    group_id INT REFERENCES groups(id) ON DELETE CASCADE,
    lesson_id INT REFERENCES lessons(id) ON DELETE CASCADE,
    PRIMARY KEY (group_id, lesson_id)
);

--СОЗДАНИЕ VIEW

--Создание view группы
CREATE OR REPLACE VIEW groups_view AS
SELECT
    g.id AS g_id,
    g.name AS g_name,
    f.id AS f_id,
    f.name AS f_name,
    g.course AS course,
    et.id AS et_id,
    et.name AS et_name
FROM
    groups g
LEFT JOIN
    faculties f ON g.faculty_id = f.id
LEFT JOIN
    education_types et ON g.education_type_id = et.id;

--Создание view учителя
CREATE OR REPLACE VIEW teachers_view AS
SELECT 
    t.id AS t_id,
    t.name AS t_name,
    dp.id AS dp_id,
    dp.name AS dp_name
FROM
    teachers t
LEFT JOIN
    departments dp ON t.department_id = dp.id;

