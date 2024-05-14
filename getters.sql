SELECT lesson_num 'Номер пары', start_time 'Начало пары',
       end_time 'Конец пары', lesson_types.name 'Тип занятия',
       disciplines.name 'Название дисциплины',
       rooms.building_num 'Корпус', rooms.num 'Номер аудитории',
       string_agg(groups.name, ', ') 'Группа(ы)',
       subgroup_num 'Номер подгруппы', string_agg(teachers.name, ', ') 'Преподаватель(и)',
       cathedras.name 'Кафедра', is_commission 'Это коммиссия?'
       FROM lessons
       JOIN lesson_time ON lesson_num = lesson_time.num
       JOIN lesson_types ON lesson_type_id = lesson_types.id
       JOIN disciplines ON discipline_id = disciplines.id
       JOIN rooms ON room_id = rooms.id
       join m2m_groups_lessons g_l ON lessons.id = g_l.lesson_id 
       JOIN groups ON group_id = groups.id
       JOIN m2m_teachers_lessons t_l ON lessons.id = t_l.lesson_id
       JOIN teachers ON teacher_id = teachers.id
       JOIN cathedras ON cathedra_id = cathedras.id
        GROUP BY lessons.id;

SELECT DISTINCT building_num FROM rooms ORDER BY building_num ASC;

SELECT num FROM rooms WHERE building_num = $1 ORDER BY num ASC;

SELECT lessons.id, disciplines.name, lesson_types.name, date,
       lesson_num, start_time,
       end_time, rooms.building_num, rooms.num,
       cathedras.name, is_commission
       FROM lessons
       JOIN lesson_types ON lesson_type_id = lesson_types.id
       JOIN lesson_time ON lesson_num = lesson_time.num
       JOIN disciplines ON discipline_id = disciplines.id
       JOIN rooms ON room_id = rooms.id
       JOIN cathedras ON cathedra_id = cathedras.id
       WHERE building_num = $1 AND rooms.num = $2
       AND extract('week' from date) = $3;

