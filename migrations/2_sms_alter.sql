ALTER TABLE sms
ADD priority INTEGER CHECK (priority < 2 AND priority >= 0)
