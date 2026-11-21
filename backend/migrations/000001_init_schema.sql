-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. ENUMS
-- +goose StatementBegin
DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
        CREATE TYPE user_role AS ENUM ('admin', 'school', 'teacher', 'student');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_gender') THEN
        CREATE TYPE user_gender AS ENUM ('male', 'female');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ai_status') THEN
        CREATE TYPE ai_status AS ENUM ('pending', 'processing', 'completed', 'failed');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ai_type') THEN
        CREATE TYPE ai_type AS ENUM ('text', 'image', 'video');
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'subscription_plan') THEN
        CREATE TYPE subscription_plan AS ENUM ('free', 'basic', 'ultra');
    END IF;
END $$;
-- +goose StatementEnd

-- 2. TABLES
CREATE TABLE IF NOT EXISTS schools (
    id BIGSERIAL PRIMARY KEY,
    name_kz TEXT NOT NULL,
    name_ru TEXT NOT NULL,
    region_id INTEGER NOT NULL,
    district_id INTEGER NOT NULL,
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20) NOT NULL UNIQUE,
    email VARCHAR(255) UNIQUE,
    iin VARCHAR(12) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    birth_year INTEGER NOT NULL,
    gender user_gender NOT NULL,
    role user_role NOT NULL,
    school_id BIGINT REFERENCES schools(id),
    google_id VARCHAR(255) UNIQUE,
    
    -- Subscriptions
    subscription_plan subscription_plan NOT NULL DEFAULT 'free',
    tokens_used INTEGER NOT NULL DEFAULT 0,
    tokens_limit INTEGER NOT NULL DEFAULT 100,
    students_count INTEGER NOT NULL DEFAULT 0,
    students_limit INTEGER NOT NULL DEFAULT 10,
    subscription_started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    subscription_expires_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS ai_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type ai_type NOT NULL,
    prompt TEXT NOT NULL,
    response TEXT,
    status ai_status NOT NULL DEFAULT 'pending',
    error_msg TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS token_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tokens_consumed INTEGER NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sender_id UUID NULL,
    sender_name VARCHAR(255) NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS push_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform VARCHAR(20) NOT NULL,
    device_name VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(20),
    email VARCHAR(255),
    code VARCHAR(6) NOT NULL,
    type VARCHAR(20) NOT NULL,
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    attempts INTEGER NOT NULL DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    verified_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS verification_blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    identifier VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    blocked_until TIMESTAMP WITH TIME ZONE,
    is_permanently_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS classes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id BIGINT REFERENCES schools(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS class_students (
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (class_id, student_id)
);

CREATE TABLE IF NOT EXISTS class_invites (
    id BIGSERIAL PRIMARY KEY,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL UNIQUE,
    usage_limit INT NOT NULL DEFAULT 1,
    usage_count INT NOT NULL DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID,
    role VARCHAR(20),
    action VARCHAR(50) NOT NULL,
    entity VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100),
    payload JSONB,
    source VARCHAR(20),
    request_id VARCHAR(100),
    ip_address VARCHAR(45),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- KMJ (Short Term Plans)
CREATE TABLE IF NOT EXISTS national_values (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    month_number INT NOT NULL CHECK (month_number >= 1 AND month_number <= 12),
    name_kaz VARCHAR(255) NOT NULL,
    name_rus VARCHAR(255) NOT NULL,
    name_eng VARCHAR(255) NOT NULL,
    description_kaz TEXT,
    description_rus TEXT,
    description_eng TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT unique_month_value UNIQUE (month_number, name_kaz, deleted_at)
);

CREATE TABLE IF NOT EXISTS kmj_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id UUID REFERENCES users(id) ON DELETE SET NULL,
    national_value_id UUID REFERENCES national_values(id),
    subject VARCHAR(255) NOT NULL,
    grade INTEGER NOT NULL CHECK (grade >= 1 AND grade <= 12),
    lesson_number INTEGER NOT NULL,
    lesson_topic VARCHAR(500) NOT NULL,
    teacher_name VARCHAR(255) NOT NULL,
    lesson_date DATE,
    content_lang VARCHAR(10) NOT NULL DEFAULT 'kaz',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    generation_task_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kmj_cells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES kmj_plans(id) ON DELETE CASCADE,
    section VARCHAR(100) NOT NULL,
    cell_key VARCHAR(255) NOT NULL,
    cell_value TEXT NOT NULL,
    stage_column VARCHAR(100),
    is_ai_generated BOOLEAN DEFAULT false,
    is_edited BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kmj_generation_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES kmj_plans(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id UUID REFERENCES users(id) ON DELETE SET NULL,
    prompt TEXT NOT NULL,
    language VARCHAR(10) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    retry_count INTEGER DEFAULT 0,
    error_message TEXT,
    ai_response_id UUID,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS teaching_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_kaz VARCHAR(255) NOT NULL UNIQUE,
    name_rus VARCHAR(255) NOT NULL,
    name_eng VARCHAR(255) NOT NULL,
    type_kaz VARCHAR(100),
    purpose_kaz TEXT,
    purpose_rus TEXT,
    purpose_eng TEXT,
    effectiveness_kaz TEXT,
    effectiveness_rus TEXT,
    effectiveness_eng TEXT,
    application_kaz TEXT,
    application_rus TEXT,
    application_eng TEXT,
    modules TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kmj_plan_methods (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id UUID NOT NULL REFERENCES kmj_plans(id) ON DELETE CASCADE,
    method_id UUID NOT NULL REFERENCES teaching_methods(id),
    sort_order INT NOT NULL CHECK (sort_order >= 1 AND sort_order <= 4),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(plan_id, method_id),
    UNIQUE(plan_id, sort_order)
);

-- 3. INDEXES
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_iin ON users(iin);
CREATE INDEX IF NOT EXISTS idx_users_school_id ON users(school_id);
CREATE INDEX IF NOT EXISTS idx_users_subscription ON users(subscription_plan);
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);

CREATE INDEX IF NOT EXISTS idx_ai_requests_user ON ai_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_requests_status ON ai_requests(status);

CREATE INDEX IF NOT EXISTS idx_token_usage_user ON token_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_token_usage_created ON token_usage(created_at);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at);

CREATE INDEX IF NOT EXISTS idx_push_tokens_user ON push_tokens(user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_push_tokens_unique ON push_tokens(user_id, token);

CREATE INDEX IF NOT EXISTS idx_verification_codes_phone ON verification_codes(phone);
CREATE INDEX IF NOT EXISTS idx_verification_codes_email ON verification_codes(email);
CREATE UNIQUE INDEX IF NOT EXISTS idx_verification_blocks_identifier ON verification_blocks(identifier);

CREATE INDEX IF NOT EXISTS idx_classes_teacher_id ON classes(teacher_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);

CREATE INDEX IF NOT EXISTS idx_kmj_plans_owner ON kmj_plans(owner_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kmj_plans_status ON kmj_plans(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kmj_plans_created ON kmj_plans(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_kmj_cells_plan ON kmj_cells(plan_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_kmj_cells_unique_identity ON kmj_cells (plan_id, section, cell_key, (COALESCE(stage_column, '')));

CREATE INDEX IF NOT EXISTS idx_national_values_month ON national_values(month_number) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teaching_methods_deleted ON teaching_methods(deleted_at);

-- 4. SEED DATA
-- Default Admin
INSERT INTO users (
    phone, iin, email, password_hash, first_name, last_name, birth_year, gender, role
) VALUES (
    '77088710495', '123456789012', 'orazbakhov@bk.ru', '$2a$10$zYNLZs6izQZOl79sBs9Zq.oAwcy84oGE30gYDauh4nTzq7V.NhY.G', 'Айбек', 'Оразбахов', 1990, 'male', 'admin'
) ON CONFLICT (email) DO NOTHING;

-- National Values (All 12 Months)
INSERT INTO national_values (month_number, name_kaz, name_rus, name_eng, description_kaz, description_rus, description_eng) VALUES
(1, 'Заң және тәртіп', 'Закон и порядок', 'Law and Order', 'Заң үстемдігін қамтамасыз ету, құқықтық мәдениетті көтеру', 'Обеспечение верховенства закона, повышение правовой культуры', 'Ensuring the rule of law, raising legal culture'),
(2, 'Жасампаздық және жаңашылдық', 'Созидание и новаторство', 'Creativity and Innovation', 'Жаңа идеялар мен шығармашылыққа ұмтылу, өзгерістерге ашық болу', 'Стремление к новым идеям и творчеству, открытость к переменам', 'Striving for new ideas and creativity, openness to change'),
(3, 'Тәуелсіздік және отаншылдық', 'Независимость и патриотизм', 'Independence and Patriotism', 'Отанға деген сүйіспеншілік, мемлекеттік мүддені қорғау', 'Любовь к Родине, защита государственных интересов', 'Love for the Motherland, protection of state interests'),
(4, 'Еңбекқорлық және кәсіби біліктілік', 'Трудолюбие и профессионализм', 'Hard work and Professionalism', 'Еңбекті бағалау, өз ісінің шебері болуға ұмтылу', 'Ценность труда, стремление быть мастером своего дела', 'Valuing labor, striving to be a master of one''s craft'),
(5, 'Бірлік және ынтымақ', 'Единство и солидарность', 'Unity and Solidarity', 'Татулық, өзара қолдау және келісім', 'Согласие, взаимная поддержка и договоренность', 'Harmony, mutual support and agreement'),
(6, 'Жасампаздық және жаңашылдық', 'Созидание и новаторство', 'Creativity and Innovation', 'Жаңашылдыққа ұмтылу, ғылым мен білімге құштарлық', 'Стремление к новаторству, тяга к знаниям и науке', 'Striving for innovation, passion for science and knowledge'),
(7, 'Заң және тәртіп', 'Закон и порядок', 'Law and Order', 'Қоғамдық тәртіпті сақтау, заң талаптарын орындау', 'Соблюдение общественного порядка, выполнение требований закона', 'Maintaining public order, complying with legal requirements'),
(8, 'Заң және тәртіп', 'Закон и порядок', 'Law and Order', 'Құқықтық сананы арттыру, қауіпсіз қоғам құру', 'Повышение правового сознания, создание безопасного общества', 'Raising legal awareness, creating a safe society'),
(9, 'Еңбекқорлық және кәсіби біліктілік', 'Трудолюбие и профессиональная компетентность', 'Diligence and professional competence', 'Еңбексүйгіштік, кәсіби шеберлік пен дағдыларды дамыту', 'Развитие трудолюбия, профессионального мастерства и навыков', 'Development of diligence, professional skills and competencies'),
(10, 'Тәуелсіздік және отаншылдық', 'Независимость и патриотизм', 'Independence and patriotism', 'Отанға деген сүйіспеншілік, тәуелсіздікті бағалау', 'Любовь к Родине, ценность независимости', 'Love for homeland, value of independence'),
(11, 'Әділдік және жауапкершілік', 'Справедливость и ответственность', 'Justice and responsibility', 'Әділеттілік, жауапкершілікті сезіну', 'Справедливость, чувство ответственности', 'Fairness, sense of responsibility'),
(12, 'Бірлік және ынтымақ', 'Единство и сотрудничество', 'Unity and cooperation', 'Ұжымдық жұмыс, бірлік пен ынтымақтастық', 'Коллективная работа, единство и сотрудничество', 'Unity and cooperation')
ON CONFLICT DO NOTHING;

-- Teaching Methods (All 50)
INSERT INTO teaching_methods (name_kaz, name_rus, name_eng, type_kaz, purpose_kaz, purpose_rus, purpose_eng, modules) VALUES
('«Миға шабуыл»', 'Мозговой штурм', 'Brainstorming', 'әдіс', 'Ақпарат туралы бар білгендерін жазбаша немесе ауызша ой салу.', 'Высказывание идей по заданной теме устно или письменно', 'Expressing ideas on a given topic orally or in writing', 'СТО, Диалогтік оқыту'),
('«Шаттық шеңбері»', 'Круг радости', 'Circle of Joy', 'сергіту сәті', 'Идея немесе тілек білдіру, тыңдау дағдыларын дамыту.', 'Развитие навыков выражения идей и пожеланий, слушания', 'Developing skills of expressing ideas and wishes, listening', 'ЖАДА'),
('«Ақылдың алты ойлау қалпағы»', 'Шесть шляп мышления', 'Six Thinking Hats', 'әдіс', 'Сыни тұрғыдан ойлауын дамыту.', 'Развитие критического мышления', 'Developing critical thinking', 'СТО, Кері байланыс'),
('«Гүлмен тілек»', 'Пожелание с цветком', 'Flower Wish', 'сергіту сәті', 'Жағымды ахуал қалыптастыру.', 'Создание позитивной атмосферы', 'Creating positive atmosphere', 'ЖАДА'),
('«Сұраққа жетіп ал»', 'Догони вопрос', 'Catch the Question', 'әдіс', 'Өз бетінше жұмыс жасау дағдысын қалыптастыру.', 'Формирование навыков самостоятельной работы', 'Forming independent work skills', 'СТО, диалогтік оқыту'),
('«Галереяны шарлау»', 'Прогулка по галерее', 'Gallery Walk', 'әдіс', 'Постер бойынша қорытынды жұмыс немесе өнімді ұсыну.', 'Просмотр, оценка и обсуждение групповых работ', 'Viewing, evaluating, and discussing group work', 'СТО, диалогтік оқыту'),
('«Биопоэма»', 'Биопоэма', 'Biopoem', 'әдіс', 'Бер-бірімен танысу, ақпарат жинақтау.', 'Знакомство, сбор информации', 'Intro, info collection', 'ЖАДА'),
('«Менің есімімнің тарихы»', 'История моего имени', 'History of My Name', 'әдіс', 'Оқушыларды таныстыру, құндылықтарды қалыптастыру.', 'Знакомство, формирование ценностей', 'Intro, values formation', 'ЖАДА'),
('«Джигсо»', 'Джигсо', 'Jigsaw', 'әдіс', 'Тақырыпты топ ішінде талқылау.', 'Обсуждение темы в группе', 'Group discussion', 'СТО, Көшбасшылық'),
('«INSERT»', 'ИНСЕРТ', 'INSERT', 'әдіс', 'Оқылған мәліметті түртіп алу жүйесі.', 'Систематизация при чтении', 'Systematization while reading', 'СТО, Бағалау'),
('«Кластер»', 'Кластер', 'Cluster', 'әдіс', 'Идеялар арасындағы байланыстарды жинақтау.', 'Сбор связей между идеями', 'Collecting connections', 'СТО'),
('«Фишбоун»', 'Фишбоун', 'Fishbone', 'әдіс', 'Жаңа мәліметті сұрақ-жауап арқылы жинақтау.', 'Выявление причинно-следственных связей', 'Cause-effect analysis', 'СТО, диалогтік оқыту'),
('«Өрмекші»', 'Паук', 'Spider', 'әдіс', 'Түйінді ойларды бекіту.', 'Закрепление ключевых идей', 'Consolidating key ideas', 'СТО'),
('«Үштік» (Ойлан, жұптас, бөліс)', 'Тройка (Думай, в паре, делись)', 'Think-Pair-Share', 'әдіс', 'Идеяларды жазу және бөлісу.', 'Индивидуальное обдумывание и обсуждение', 'Individual thinking and sharing', 'СТО, диалогтік оқыту'),
('«Эссе» немесе «Еркін әңгіме»', 'Эссе или Свободный рассказ', 'Essay or Free Talk', 'әдіс', 'Өз ойына талдау жасау.', 'Анализ собственных мыслей', 'Analyzing own thoughts', 'Бағалау'),
('«Қар кесегі»', 'Снежный ком', 'Snowball', 'әдіс', 'Ұжымдасып ой бөлісу.', 'Коллективный обмен мыслями', 'Collective sharing', 'СТО, диалогтік оқыту'),
('«Заряд алу»', 'Зарядка', 'Recharge', 'сергіту сәті', 'Ұқыптылық пен сезімталдыққа үйрету.', 'Обучение внимательности и чувствительности', 'Training care and sensitivity', 'ЖАДА'),
('«Бұзық телефон»', 'Испорченный телефон', 'Broken Telephone', 'сергіту сәті', 'Дұрыс тыңдалымға бағыттау.', 'Направление на правильное слушание', 'Directing to proper listening', 'ЖАДА'),
('«Бэкроним»', 'Бэкроним', 'Backronym', 'әдіс', 'Логикалық ойлауды дамыту.', 'Развитие логического мышления', 'Developing logical thinking', 'СТО'),
('SMART-мақсаттар', 'SMART-цели', 'SMART Goals', 'әдіс', 'Нақты мақсаттар құру.', 'Построение конкретных целей', 'Building specific goals', 'СТО'),
('«Мюллердің ғажайып саны»', 'Число Мюллера', 'Muller Number', 'әдіс', 'Сабақ мақсатын нақты қою.', 'Четкая постановка целей урока', 'Clear lesson goal setting', 'СТО'),
('«Константинополь»', 'Константинополь', 'Constantinople', 'әдіс', 'Ойлау қабілетін дамыту.', 'Развитие мыслительных способностей', 'Developing thinking abilities', 'СТО'),
('SWOT - талдау', 'SWOT - анализ', 'SWOT - analysis', 'әдіс', 'Мәселені талдау және салыстыру.', 'Анализ и сравнение проблемы', 'Analyzing and comparing problem', 'СТО, Кері байланыс'),
('«Зерттеушілік конференциясы»', 'Исследовательская конференция', 'Research Conference', 'әдіс', 'Рефлексия жасауды дамыту.', 'Развитие рефлексии', 'Developing reflection', 'СТО, Көшбасшылық'),
('«Кейс – стади»', 'Кейс – стади', 'Case Study', 'әдіс', 'Жағдаятты талдау және болжау.', 'Анализ и прогнозирование ситуаций', 'Analysis and forecasting situations', 'СТО, АКТ'),
('«Дауыстап ойлау»', 'Мысли вслух', 'Thinking Aloud', 'әдіс', 'Оқи отырып ой түю.', 'Размышление во время чтения', 'Thinking while reading', 'СТО, диалогтік оқыту'),
('«Шабадан, еттартқыш, қоқыс жәшігі»', 'Чемодан, мясорубка, корзина', 'Suitcase, Mincer, Bin', 'кері байланыс', 'Алған білімін саралау.', 'Сортировка полученных знаний', 'Sorting acquired knowledge', 'Кері байланыс'),
('«Ыстық орындық»', 'Горячий стул', 'Hot Seat', 'әдіс', 'Оқушы белсенділігін арттыру.', 'Повышение активности учащихся', 'Increasing student activity', 'СТО, диалогтік оқыту'),
('«Бір минуттық әңгіме»', 'Минутка разговора', 'One Minute Talk', 'әдіс', 'Бір-бірінен ақпарат алу.', 'Получение информации друг от друга', 'Getting info from each other', 'СТО, диалогтік оқыту'),
('«Кинометафора»', 'Кинометафора', 'Movie Metaphor', 'әдіс', 'Ой-өрісін дамыту.', 'Развитие кругозора', 'Developing horizons', 'СТО, АКТ'),
('«Қара жәшік» ішіндегі жұмыс', 'Работа в черном ящике', 'Working in Black Box', 'әдіс', 'Өзін-өзі бағалауды дамыту.', 'Развитие самооценки', 'Developing self-assessment', 'ОүБ'),
('«Бес саусақ»', 'Пять пальцев', 'Five Fingers', 'кері байланыс', 'Бір-біріне бірігуге ықпал ету.', 'Содействие объединению', 'Promoting unification', 'Кері байланыс'),
('«Ашық микрофон»', 'Открытый микрофон', 'Open Mic', 'кері байланыс', 'Жылдам ойлауға дағдыландыру.', 'Обучение быстрому мышлению', 'Training fast thinking', 'СТО, Кері байланыс'),
('«Шахмат»', 'Шахматы', 'Chess', 'кері байланыс', 'Кері байланысты жазбаша жүргізу.', 'Письменная обратная связь', 'Written feedback', 'СТО, Кері байланыс'),
('«Төрттік» (әке, ана, бала, дос)', 'Четверка (отец, мать, ребенок, друг)', 'Foursome', 'әдіс', 'Бір-біріне жақындастыру.', 'Сближение друг с другом', 'Bringing together', 'ЖАДА'),
('«Ханшайым, айдаһар, Ханзада»', 'Принцесса, дракон, принц', 'Princess, Dragon, Prince', 'сергіту сәті', 'Шапшаңдыққа баулу.', 'Обучение быстроте', 'Training speed', 'ЖАДА'),
('«Сұрақ ілмегі»', 'Петля вопросов', 'Question Loop', 'кері байланыс', 'Ойды ашық білдіру.', 'Открытое выражение мыслей', 'Open expression of thoughts', 'СТО, Кері байланыс'),
('«Екі дұрыс, бір бұрыс»', 'Два правды, одна ложь', 'Two Truths, One Lie', 'әдіс', 'Ойлау қабілетін дамыту.', 'Развитие мыслительных способностей', 'Developing thinking abilities', 'СТО'),
('«Жалғасын тап»', 'Найди продолжение', 'Find the Sequence', 'әдіс', 'Жылдам ой қорыту.', 'Быстрое обобщение мыслей', 'Fast thought generalization', 'СТО'),
('«Буклет»', 'Буклет', 'Booklet', 'әдіс', 'Ақпаратты жинақтау.', 'Сбор информации', 'Information collection', 'СТО'),
('«Тормоз»', 'Тормоз', 'Brake', 'сергіту сәті', 'Тыңдалым мен шапшаңдық.', 'Слушание и быстрота', 'Listening and speed', 'ЖАДА'),
('«Шеринг»', 'Шеринг', 'Sharing', 'кері байланыс', 'Ой бөлісу.', 'Обмен мыслями', 'Sharing thoughts', 'ЖАДА, Кері байланыс'),
('«Ит па, мысық па?»', 'Собака или кошка?', 'Dog or Cat?', 'сергіту сәті', 'Зейін мен шапшаңдық.', 'Внимание и быстрота', 'Attention and speed', 'ЖАДА'),
('«Эйлер шеңбері»', 'Круги Эйлера', 'Euler Circles', 'әдіс', 'Идеяларды жинақтап саралау.', 'Сбор и дифференциация идей', 'Collecting and sorting ideas', 'СТО'),
('«МИКС» жаттығуы', 'Упражнение МИКС', 'MIX Exercise', 'кері байланыс', 'Идеяларды түйіндеу.', 'Обобщение идей', 'Summarizing ideas', 'ЖАДА, Кері байланыс'),
('«Тепе-теңдік дөңгелегі»', 'Колесо баланса', 'Balance Wheel', 'кері байланыс', 'Қажеттіліктерді анықтау.', 'Определение потребностей', 'Identifying needs', 'СТО, Кері байланыс'),
('«Кувёртка»', 'Кувертка', 'Kuvertka', 'әдіс', 'Идеяларды жазбаша жинақтау.', 'Письменный сбор идей', 'Written idea collection', 'Диалогтік оқыту'),
('«Желпуіш» (Веер)', 'Веер', 'Fan', 'кері байланыс', 'Жазбаша кері байланыс.', 'Письменная обратная связь', 'Written feedback', 'СТО, Диалготік оқыту'),
('«Ұшақ»', 'Самолетик', 'Plane', 'кері байланыс', 'Идеяларды жинақтау.', 'Сбор идей', 'Idea collection', 'Кері байланыс'),
('«Жасырын сөздер»', 'Скрытые слова', 'Hidden Words', 'кері байланыс', 'Қорытындылау және кері байланыс.', 'Итоги и обратная связь', 'Summary and feedback', 'Кері байланыс')
ON CONFLICT (name_kaz) DO NOTHING;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
DROP TABLE IF EXISTS kmj_plan_methods CASCADE;
DROP TABLE IF EXISTS teaching_methods CASCADE;
DROP TABLE IF EXISTS kmj_generation_tasks CASCADE;
DROP TABLE IF EXISTS kmj_cells CASCADE;
DROP TABLE IF EXISTS kmj_plans CASCADE;
DROP TABLE IF EXISTS national_values CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS class_invites CASCADE;
DROP TABLE IF EXISTS class_students CASCADE;
DROP TABLE IF EXISTS classes CASCADE;
DROP TABLE IF EXISTS verification_blocks CASCADE;
DROP TABLE IF EXISTS verification_codes CASCADE;
DROP TABLE IF EXISTS push_tokens CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS token_usage CASCADE;
DROP TABLE IF EXISTS ai_requests CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS schools CASCADE;
DROP TYPE IF EXISTS subscription_plan CASCADE;
DROP TYPE IF EXISTS ai_type CASCADE;
DROP TYPE IF EXISTS ai_status CASCADE;
DROP TYPE IF EXISTS user_gender CASCADE;
DROP TYPE IF EXISTS user_role CASCADE;
