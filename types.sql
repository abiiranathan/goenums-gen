/*
This file was not generated by pg_dump command.
The comments will be safely escaped.
*/

-- Tracks the status of a patient on the Theatre list.
CREATE TYPE operation_status AS ENUM(
  'PENDING',
  'ON GOING',
  'COMPLETED',
  'POSTPONED',
  'CANCELLED'
);