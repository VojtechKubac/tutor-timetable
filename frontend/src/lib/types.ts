export interface Teacher {
	id: string;
	email: string;
	name: string;
	created_at: string;
}

export interface TeacherSettings {
	teacher_id: string;
	working_start: string; // "HH:MM"
	working_end: string;   // "HH:MM"
	lesson_duration_minutes: number;
	max_gap_minutes: number;
	max_consecutive_lessons: number;
	break_after_n_lessons: number; // 0 = disabled
	break_duration_minutes: number;
	locale: string;
}

export interface Student {
	id: string;
	teacher_id: string;
	name: string;
	email?: string;
	notes?: string;
	created_at: string;
}

export interface AvailabilitySlot {
	id?: string;
	owner_type?: string;
	owner_id?: string;
	day_of_week: number; // 0=Mon … 6=Sun
	start_time: string;  // "HH:MM"
	end_time: string;    // "HH:MM"
}

export interface Lesson {
	id: string;
	teacher_id: string;
	student_id: string;
	student_name?: string;
	day_of_week: number;
	start_time: string;
	end_time: string;
	is_pinned: boolean;
	created_at: string;
	updated_at: string;
}
