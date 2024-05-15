const overlay = document.getElementById('overlay');
const modal = document.getElementById('details');
const closeButton = document.getElementById('close-button');
const timetable = ['08:30<br>10:00', '10:10<br>11:40',
'11:50<br>13:20', '14:00<br>15:30',
'15:40<br>17:10', '17:20<br>18:50',
'18:55<br>20:25', '20:30<br>22:00']

function getDayOfWeek(dateString) {
    // Создаем объект Date из строки с датой и таймзоной
    const date = new Date(dateString);

    // Массив с названиями дней недели
    const daysOfWeek = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];

    // Получаем номер дня недели (0 - воскресенье, 1 - понедельник и т.д.)
    const dayIndex = date.getDay();

    // Возвращаем название дня недели из массива
    return daysOfWeek[dayIndex];
}

function changeWeek(num) {  
    weekNumSelect.value = Number(weekNumSelect.value) +num;
}

function setCookie(name,value,days) {
var expires = "";
if (days) {
    var date = new Date();
    date.setTime(date.getTime() + (days*24*60*60*1000));
    expires = "; expires=" + date.toUTCString();
}
document.cookie = name + "=" + (value || "")  + expires + "; path=/";
}

function getCookie(name) {
var nameEQ = name + "=";
var ca = document.cookie.split(';');
for(var i=0;i < ca.length;i++) {
    var c = ca[i];
    while (c.charAt(0)==' ') c = c.substring(1,c.length);
    if (c.indexOf(nameEQ) == 0) return c.substring(nameEQ.length,c.length);
}
return null;
}

function eraseCookie(name) {   
document.cookie = name +'=; Path=/; Expires=Thu, 01 Jan 1970 00:00:01 GMT;';
}

// Функция получения номера недели по РЭУ
function getCurrentReaWeek() {
const currentDate = new Date();
const januaryFirst = 
    new Date(currentDate.getFullYear(), 0, 1);
const daysToNextMonday = 
    (januaryFirst.getDay() === 1) ? 0 : 
    (7 - januaryFirst.getDay()) % 7;
const nextMonday = 
    new Date(currentDate.getFullYear(), 0, 
    januaryFirst.getDate() + daysToNextMonday);

const weekNum = (currentDate < nextMonday) ? 52 : 
(currentDate > nextMonday ? Math.ceil(
(currentDate - nextMonday) / (24 * 3600 * 1000) / 7) : 1);
return (weekNum - 34 <= 0) ? weekNum + 18 : weekNum - 34
}      


// дата в дд.мм.гг
function formatDate(dateString) {
let date = new Date(dateString);
let day = date.getDate().toString().padStart(2, '0'); // Получаем день месяца с добавлением ведущего нуля
let month = (date.getMonth() + 1).toString().padStart(2, '0'); // Получаем месяц с добавлением ведущего нуля
let year = date.getFullYear().toString().slice(-2); // Получаем последние две цифры года

return `${day}.${month}.${year}`;
}

// Получения списка дат недели
function getWeekDates(weekNum) {
    let currentDate = new Date(); // Текущая дата
    let currentYear = currentDate.getFullYear(); // Текущий год
    let januaryFirst = new Date(currentYear, 0, 1); // Дата 1 января текущего года
    let januaryFirstDay = januaryFirst.getDay(); // День недели, на который приходится 1 января
    let firstMondayDate = new Date(januaryFirst); // Первый понедельник года
    firstMondayDate.setDate(firstMondayDate.getDate() + (1 - januaryFirstDay) % 7); // Установка на первый понедельник

    // Получаем дату первого дня недели с учетом номера недели
    let weekStartDate = new Date(firstMondayDate);
    weekStartDate.setDate(firstMondayDate.getDate() + (weekNum - 1) * 7); // Установка на первый день недели

    let weekDates = []; // Массив для хранения дат недели

    // Заполняем массив датами недели
    for (let i = 0; i < 7; i++) {
        let date = new Date(weekStartDate);
        date.setDate(weekStartDate.getDate() + i); // Устанавливаем дату для текущего дня недели
        weekDates.push(date); // Добавляем дату в массив в формате дд.мм.гггг
    }

return weekDates;
}

// Приведение даты из рэувского формата в рэушный
function toISOWeekNum(weekNum) {
weekNum = Number(weekNum)
weekNum += 34
if (weekNum > 52) {
    weekNum = weekNum - 34 - 18
}
return weekNum
}

// Получить список аудиторий
async function getRoomsByBuildingNum(buildingNum) {
    try {
        const response = await fetch(`api/rooms/${buildingNum}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null; 
    }
}

// Получить расписание по аудитории
async function getScheduleForRoom(weekNum, roomId) {
    try {
        const response = await fetch(`api/schedule/${weekNum}/room/${roomId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

// Получить расписание по преподавателю
async function getScheduleForTeacher(weekNum, teacherId) {
    try {
        const response = await fetch(`api/schedule/${weekNum}/teacher/${teacherId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

// Получить расписание по группе
async function getScheduleForGroup(weekNum, groupId) {
    try {
        const response = await fetch(`api/schedule/${weekNum}/group/${groupId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getCathedras(){
    try {
        const response = await fetch(`api/teachers/getCathedras`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getTeachers(cathedraId){
    try {
        const response = await fetch(`api/teachers/getTeachers?cathedraId=${cathedraId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getFaculties() {
    try {
        const response = await fetch(`api/groups/getFaculties`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getCourses(facultyId) {
    try {
        const response = await fetch(`api/groups/getCourses?facultyId=${facultyId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getEducationTypes(facultyId, course){
    try {
        const response = await fetch(`api/groups/getEducationTypes?facultyId=${facultyId}&course=${course}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function getGroups(facultyId, course, educationTypeId) {
    try {
        const response = await fetch(`api/groups/getGroups?facultyId=${facultyId}&course=${course}&educationTypeId=${educationTypeId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

async function search(someString){
    try {
        const response = await fetch(`api/search?q=${teacherId}`);
        if (!response.ok) {
            throw new Error('Ошибка при получении данных');
        }
        return await response.json();
    } catch (error) {
        console.error("Ошибка при запросе к API:", error);
        return null;
    }
}

function createTimeslot(lessonType) {
    let timeslot = document.createElement("div");

    timeslot.classList.add("timeslot");
    switch (lessonType){
        case 'Практическое занятие':
            timeslot.classList.add('practice');
        break;
        case 'Лекция':
            timeslot.classList.add('lecture');
        break;
        case 'Лабораторная работа':
            timeslot.classList.add('lab');
        break;
        case "empty":
            timeslot.classList.add("empty");
            break;  
        default:
            timeslot.classList.add('exam');
        break;
    }
    return timeslot
}


function createTimeinfo(lessonNum) {
    let timeinfo = document.createElement("div");
    timeinfo.classList.add("timeinfo");
    timeinfo.classList.add(lessonNum)
    timeinfo.innerHTML = `${lessonNum} пара<br>${timetable[lessonNum-1]}`
    return timeinfo
}


function createLessonInfo(lesson) {
    let lessonInfo = document.createElement("div");
    lessonInfo.classList.add("lessoninfo")
    lessonInfo.innerHTML = `${lesson.Discipline.Name}<br><i>${lesson.LessonType.Name}`
    if (lesson.IsCommission == true) {
        lessonInfo.innerHTML += "  - Комиссия"
    }
    lessonInfo.innerHTML += "</i><br>"
    // Если вызвано при получении расписания аудитории
    if (lesson.SubgroupNum != 0){
        lessonInfo.innerHTML += "<b>+ подгруппы</b>"
    }
    else if (lesson.Room.Id == 0) {
        lessonInfo.innerHTML += `${lesson.Teachers[0].Name}`
        if (lesson.Teachers.length > 1){
            lessonInfo.innerHTML += " и др."
        }
    }
    return lessonInfo
}


function createDayTitle(date) {
    let dayTitle = document.createElement("div");
    dayTitle.classList.add("dayTitle");
    let d = new Date(date);
    let today = new Date();
    if (d.getDate() == today.getDate()) {
        dayTitle.id = "today";
    }
    dayTitle.innerHTML = `${getDayOfWeek(date)}, ${formatDate(date)}`;
    return dayTitle;
}

// Заполнить данные урока
function setLesson(lesson) {
    let dayElement = document.getElementById(formatDate(lesson.Date));

    let EmptyDays = dayElement.getElementsByClassName("emptyDay");
    if (EmptyDays.length != 0){
        EmptyDays[0].remove();
    }
    // генерация пустых пар
    for (let i = 1; i < lesson.LessonNum; i++) {
        let timeslots = dayElement.getElementsByClassName(String(i));
        if (timeslots.length == 0) {
            let timeslot = createTimeslot("empty");

            let timeinfo = createTimeinfo(i)

            let lessonInfo = document.createElement("div");
            lessonInfo.classList.add("lessoninfo")
            lessonInfo.innerHTML = ""

            timeslot.appendChild(timeinfo)
            timeslot.appendChild(lessonInfo)
            dayElement.appendChild(timeslot)
        } 
    }
    

    let timeslots = dayElement.getElementsByClassName(String(lesson.LessonNum));
    
    if (timeslots.length == 0){
        let timeslot = createTimeslot(lesson.LessonType.Name);
        timeslot.add
        let timeinfo = createTimeinfo(lesson.LessonNum)
        let lessonInfo = createLessonInfo(lesson)
        timeslot.appendChild(timeinfo)
        timeslot.appendChild(lessonInfo)
        dayElement.appendChild(timeslot)
        timeslot.addEventListener("click", clearDetails)
        timeslot.addEventListener("click", openModal)
        timeslot.addEventListener("click", function(event) {setDetails(lesson);})
    }
    else {
        let check = timeslots[0].parentElement
        console.log(check);
        check.addEventListener("click", function(event) {setDetails(lesson);})
    }
}

// Сгенерировать выбор недели.
function generateOptions() {
    const currentWeek = getCurrentReaWeek();
    const weekNumSelect = document.getElementById("weekNumSelect");
    if (weekNumSelect) {
        for (let i = 1; i <= 52; i++) {
            const option = document.createElement("option");
            option.value = i;
            option.textContent = i;
            if (i === Number(currentWeek)) {
                option.selected = true;
            }
            weekNumSelect.appendChild(option);
        }
    }
}

// Заполнить select
async function setRoomList() {
    let buildingNum = buildingNums.selectedOptions[0].value;
    let rooms = await getRoomsByBuildingNum(buildingNum);

    roomNums.innerHTML = '';

     // Добавление пустого элемента по умолчанию
     const emptyOption = document.createElement("option");
     emptyOption.value = '';
     emptyOption.disabled = true;
     emptyOption.hidden = true;
     emptyOption.selected = true;
     emptyOption.textContent = 'Выберите аудиторию:';
     roomNums.appendChild(emptyOption);
    
    for (const room of rooms) {
        const option = document.createElement("option");
        option.value = room.Id;
        option.textContent = room.Num;
        roomNums.appendChild(option)
    }
}

function clearDetails(){
    let lessons = modal.getElementsByClassName("lesson");
    while (lessons.length>0 ){
        lessons[0].remove();
    }
}

// Заполнить подробное расписание 
function setDetails(lesson){
    let lessonElement = document.createElement("div");
    lessonElement.classList.add("lesson");
    let lessonHeader = document.createElement("div");
    lessonHeader.classList.add("lessonHeader");
    let disciplineName = document.createElement("h3");
    disciplineName.innerHTML = lesson.Discipline.Name;
    let lessonType = document.createElement("b");
    lessonType.innerHTML = lesson.LessonType.Name;
    let id = document.createElement ("i");
    id.innerHTML = `#ID: ${lesson.Id}`;
    lessonHeader.appendChild(disciplineName);
    lessonHeader.appendChild(lessonType);
    lessonHeader.appendChild(document.createElement("br"));
    lessonHeader.appendChild(id);
    lessonElement.appendChild(lessonHeader);    
    lessonElement.appendChild(document.createElement("br"));
    
    let time = document.createElement("div");
    time.classList.add("time");
    let date = new Date(lesson.Date)
    time.innerHTML = `${getDayOfWeek(date)}, ${formatDate(lesson.Date)}, ${lesson.LessonNum} пара`
    let place = document.createElement("div");
    place.classList.add("place");
    place.innerHTML = `Аудитория: ${lesson.Room.BuildingNum} корпус - ${lesson.Room.Num}`
    let groups = document.createElement("div");
    groups.classList.add("groups");
    groups.innerHTML = "<b>Группа(ы):</b> "
    for (let i=0; i<lesson.Groups.length;i++) {
        if (i>0){
            groups.innerHTML += ",  ";
        }
        groups.innerHTML += `${lesson.Groups[i].Name}`;
        if (lesson.SubgroupNum != 0) {
            groups.innerHTML += ` (${lesson.SubgroupNum})`;
        }
    }   
    let teachers = document.createElement("div");
    if (lesson.IsCommission) {
        teachers.innerHTML = "<b>Состав комиссии:</b> ";    
    }
    else {
        teachers.innerHTML = "<b>Преподаватель:</b> ";
    }
    teachers.classList.add("teachers");
    for (let i=0; i<lesson.Teachers.length;i++) {
        teachers.innerHTML += `<br>${lesson.Teachers[i].Name}`;
        if (lesson.Cathedra.Name != "") {
        teachers.innerHTML += `<br>(${lesson.Cathedra.Name})<br>`;
        }
    }
    lessonElement.appendChild(time);
    lessonElement.appendChild(place);
    lessonElement.appendChild(groups);
    lessonElement.appendChild(teachers);

    modal.appendChild(lessonElement);
}

// Заполнить schedule
function setSchedule(lessons) {

    const dates = getWeekDates(toISOWeekNum(weekNumSelect.value));
    for (let i = 0; i < dates.length - 1; i++) {
        let dayElement = document.createElement("div");
        dayElement.classList.add("day");
        dayElement.id = formatDate(dates[i]);
        let dayTitle = createDayTitle(dates[i])
        dayElement.appendChild(dayTitle)
        let emptyDay = document.createElement("div");
        emptyDay.classList.add("emptyDay");
        emptyDay.innerHTML = "<b>Занятий нет</b>"
        dayElement.appendChild(emptyDay);
        schedule.appendChild(dayElement);
    } 
    for (const lesson of lessons){
        setLesson(lesson)
    }
}

async function getAndSetScheduleForRoom() {
    let lessons = await getScheduleForRoom(weekNumSelect.value, roomNums.value);
    schedule.innerHTML = ''
    setSchedule(lessons);
}



function getSchedule() {
    switch (optionSelector.value){
        case'1':
        if (roomSelect.value != ""){
            getAndSetScheduleForRoom();
        }
        break;
        case '2':
        if (teacherSelect.value != ""){
            getAndSetScheduleForTeacher();
        }
        break;
        case '3':
        if (groupSelect.value != ""){
            getAndSetScheduleForGroup();
        }
        break;
    }
}

function closeModal() {
    overlay.classList.add("invisible")
}

function openModal() {
    overlay.classList.remove("invisible")
}

function suspend(element){
    element.classList.add("invisible");
}

function activate(element){
    element.classList.remove("invisible");
}

function suspendGroupFunctions(){
    groupSearch.classList.add("invisible");
    facultySelect.removeEventListener("change", setCourses);
    courseSelect.removeEventListener("change", setEducationTypes);
    educationTypeSelect.removeEventListener("change", setGroupsList);
    groupSelect.removeEventListener("change", getAndSetScheduleForGroup);
}

function suspendTeacherFunctions() {
    teacherSearch.classList.add("invisible");
    cathedraSelect.removeEventListener("change", setTeacherList);
    teacherSelect.removeEventListener("change", getAndSetScheduleForTeacher);
}

function suspendRoomFunctions() {
    roomSearch.classList.add("invisible");
    buildingNums.removeEventListener("change", setRoomList);
    roomNums.removeEventListener("change", getAndSetScheduleForRoom);
}

function activateGroupFunctions(){
    groupSearch.classList.remove("invisible");
    facultySelect.addEventListener("change", setCourses);
    courseSelect.addEventListener("change", setEducationTypes);
    educationTypeSelect.addEventListener("change", setGroupsList);
    groupSelect.addEventListener("change", getAndSetScheduleForGroup);
}

function activateTeacherFunctions() {
    teacherSearch.classList.remove("invisible");
    cathedraSelect.addEventListener("change", setTeacherList);
    teacherSelect.addEventListener("change", getAndSetScheduleForTeacher);
}

function activateRoomFunctions() {
    roomSearch.classList.remove("invisible");
    buildingNums.addEventListener("change", setRoomList);
    roomNums.addEventListener("change", getAndSetScheduleForRoom);
}

function activateForm() {
    console.log("a?")
    let newId = optionSelector.value;
    switch (newId) {
        case '1':
            suspendTeacherFunctions();
            suspendGroupFunctions();
            activateRoomFunctions();
            // Поиск по группе
            
        break;
        case '2':
            suspendGroupFunctions();
            suspendRoomFunctions();
            activateTeacherFunctions();
            
        break;
        case '3':
            suspendTeacherFunctions();
            suspendRoomFunctions();
            activateGroupFunctions();
        break;
        default:
            suspendTeacherFunctions();
            suspendGroupFunctions();
            suspendRoomFunctions();
        break;
    }
}

async function setCathedras(){
    const cathedras = await getCathedras();
    cathedraSelect.innerHTML = '';
    for (const cathedra of cathedras) {
        const option = document.createElement("option");
        option.value = cathedra.Id;
        option.textContent = cathedra.Name;
        cathedraSelect.appendChild(option);
    }
}

async function setTeacherList() {
    const teachers = await getTeachers(cathedraSelect.value);
    teacherSelect.innerHTML = '';
    for (const teacher of teachers) {
        const option = document.createElement("option");
        option.value = teacher.Id;
        option.textContent = teacher.Name;
        teacherSelect.appendChild(option);
    }
}

async function setFaculties() {
    const faculties = await getFaculties();
    facultySelect.innerHTML = '';
    for (const faculty of faculties) {
        const option = document.createElement("option")
        option.value = faculty.Id
        option.textContent = faculty.Name
        facultySelect.appendChild(option);
    }
}

async function setCourses() {
    const courses = await getCourses(facultySelect.value);
    courseSelect.innerHTML = '';
    for (const course of courses) {
        const option = document.createElement("option")
        option.value = course
        option.textContent = `${course}-й курс`
        courseSelect.appendChild(option);
    }
}

async function setEducationTypes() {
    const educationTypes = await getEducationTypes(facultySelect.value, courseSelect.value);
    educationTypeSelect.innerHTML = '';
    for (const educationType of educationTypes) {
        const option = document.createElement("option")
        option.value = educationType.Id
        option.textContent = educationType.Name
        educationTypeSelect.appendChild(option);
    }
}

async function setGroupsList() {
    const groups = await getGroups(facultySelect.value, courseSelect.value, educationTypeSelect.value);
    groupSelect.innerHTML = '';
    for (const group of groups) {
        const option = document.createElement("option")
        option.value = group.Id
        option.textContent = group.Name
        groupSelect.appendChild(option);
    }
}

async function getAndSetScheduleForGroup() {
    const lessons = await getScheduleForGroup(weekNumSelect.value, groupSelect.value);
    console.log(lessons);
    schedule.innerHTML = ''
    setSchedule(lessons);
}

async function getAndSetScheduleForTeacher() {
    const lessons = await getScheduleForTeacher(weekNumSelect.value, teacherSelect.value);
    schedule.innerHTML = ''
    setSchedule(lessons);
}

function setEventListeners() {
    optionSelector.addEventListener("change",  activateForm);

    // номер недели
    weekNumSelect.addEventListener("change", getSchedule);
    // Обработчик клика на кнопке закрытия
    closeButton.addEventListener('click', closeModal);

    prevWeek.addEventListener('click', function(event) {changeWeek(-1);});
    prevWeek.addEventListener('click', getSchedule);
    nextWeek.addEventListener('click', function(event){changeWeek(1);});
    nextWeek.addEventListener('click', getSchedule);
}

// Функция задания начальных значений
function init() {
    // Задаём номер недели по РЭУ и генерируем опции в выборе неделе
    generateOptions();

    setCathedras();
    setFaculties();
    // Задаём eventlistener's
    setEventListeners();
}

// Выполняем первоначальную настройку
init()



