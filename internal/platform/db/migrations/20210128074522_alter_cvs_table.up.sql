ALTER TABLE cvs ADD COLUMN full_name VARCHAR (50);
ALTER TABLE cvs ADD COLUMN phone_number VARCHAR (50);
ALTER TABLE cvs ADD COLUMN email VARCHAR (150);
ALTER TABLE cvs ADD COLUMN date_receipt_cv date;
ALTER TABLE cvs ADD COLUMN interview_method smallint;
ALTER TABLE cvs ADD COLUMN salary bigint;
ALTER TABLE cvs ADD COLUMN contact_link VARCHAR (50);

comment on column cvs.full_name is 'Name of applicant';
comment on column cvs.phone_number is 'Phone number of candidate';
comment on column cvs.email is 'Candidate email';
comment on column cvs.date_receipt_cv is 'Date receipt cv';
comment on column cvs.interview_method is 'Interview method';
comment on column cvs.salary is 'Salary';
comment on column cvs.contact_link is 'Contact link';
