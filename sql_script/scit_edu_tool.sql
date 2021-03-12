create table if not exists class_chart
(
    f_id   smallint    not null,
    s_id   smallint    not null,
    c_id   tinyint     not null,
    c_name varchar(30) not null,
    constraint class_chart
        unique (c_name, f_id, s_id, c_id)
);

create table if not exists class_schedule
(
    t_id          varchar(30)  not null,
    t_faculty     smallint     not null,
    t_specialty   smallint     not null,
    t_class       tinyint      not null,
    t_grade       smallint     not null,
    t_school_year varchar(10)  not null,
    t_semester    tinyint      not null,
    t_content     text         not null,
    t_expired     int unsigned not null,
    constraint class_schedule
        unique (t_id, t_faculty, t_specialty, t_class, t_grade, t_school_year, t_semester)
);

create table if not exists faculty_chart
(
    f_id   smallint    not null,
    f_name varchar(30) not null,
    constraint faculty_chart
        unique (f_id, f_name)
);

create table if not exists hitokoto
(
    h_id          int auto_increment
        primary key,
    h_index       int(255) default 0 not null,
    h_content     varchar(1000)      not null,
    h_type        varchar(10)        not null,
    h_from        varchar(255)       not null,
    h_from_who    varchar(255)       not null,
    h_creator     varchar(255)       not null,
    h_creator_uid int      default 0 not null,
    h_reviewer    int                not null,
    h_insert_at   int                not null,
    h_length      int                not null
);

create table if not exists sign_keys
(
    app_key    tinytext             not null,
    app_secret tinytext             not null,
    platform   tinytext             not null,
    mail       tinytext             not null,
    build      int(10)              not null,
    available  tinyint(1) default 1 null
);

create table if not exists specialty_chart
(
    s_id   smallint    not null,
    s_name varchar(30) not null,
    f_id   smallint    not null
);

create index if not exists specialty_chart
    on specialty_chart (s_name, s_id, f_id);

create table if not exists student_achieve
(
    u_id         varchar(20) not null
        primary key,
    u_faculty    smallint    not null,
    u_specialty  smallint    not null,
    u_class      tinyint     not null,
    u_grade      smallint    not null,
    a_content_01 text        null,
    a_content_02 text        null,
    a_content_03 text        null,
    a_content_04 text        null,
    a_content_05 text        null,
    a_content_06 text        null,
    a_content_07 text        null,
    a_content_08 text        null,
    a_content_09 text        null,
    a_content_10 text        null,
    a_content_11 text        null,
    a_content_12 text        null
);

create index if not exists student_achieve
    on student_achieve (u_id, u_faculty, u_specialty, u_class, u_grade);

create table if not exists user_info
(
    u_id           varchar(20)          not null
        primary key,
    u_name         tinytext             null,
    u_identify     tinyint(2) default 0 not null,
    u_level        tinyint(2) default 0 not null,
    u_faculty      smallint   default 0 not null,
    u_specialty    smallint   default 0 not null,
    u_class        tinyint    default 0 not null,
    u_grade        smallint   default 0 not null,
    u_info_expired int        default 0 not null,
    constraint user_info
        unique (u_id, u_identify, u_faculty, u_specialty, u_class, u_grade)
);

create table if not exists user_token
(
    u_id              varchar(20)             not null,
    u_password        varchar(600) default '' not null,
    u_session         varchar(30)  default '' not null,
    u_session_expired int                     not null,
    u_token_effective tinyint(1)              not null,
    constraint user_token
        unique (u_id)
);

alter table user_token
    add primary key (u_id);


