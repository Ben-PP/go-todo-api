INSERT INTO users (username, id, password_hash, created_at, is_admin) VALUES
('alice', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', '$2a$12$dTguyZXjZq1UAcZxvtnWSuGnFSgwUN0d7.fycbP964EEy9EAhNklm', '2024-01-01 10:00:00', TRUE),
('bob', '87b738aa-4fef-4334-b0d2-099614062ffd', '$2a$12$FeO/hYqBBMgAK9a03PdxweqPTLe1YvZT9mEt5sWyFwUU2JRCDeUFa', '2024-01-02 11:00:00', FALSE),
('carol', '1fe81494-9017-440a-b432-d09541d8d273', '$2a$12$dAiakREjnURifgNPfQbeSuX.PWZzrCJz9Uhz8OpZ4hqigbR5VRK6K', '2024-01-03 12:00:00', FALSE);

INSERT INTO lists (id, user_id, title, description, created_at, updated_at) VALUES
('f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Alice Work Tasks', 'Tasks related to work projects and deadlines.', '2024-01-01 10:05:00', '2024-01-01 10:05:00'),
('4f3f5366-2a49-4bd0-8501-807416d49b56', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Alice Personal Tasks', 'Personal errands and activities.', '2024-01-01 10:10:00', '2024-01-01 10:10:00'),
('0f137b9c-26fc-4c8e-924b-96edeb7d0500', '87b738aa-4fef-4334-b0d2-099614062ffd', 'Bob Shopping List', 'Groceries and household items to buy.', '2024-01-02 11:05:00', '2024-01-02 11:05:00'),
('d372481e-3fe1-4128-aba3-dee7d59ec9ce', '1fe81494-9017-440a-b432-d09541d8d273', 'Carol Reading List', 'Books and articles to read.', '2024-01-03 12:05:00', '2024-01-03 12:05:00');

-- Insert some todos for Alice, Bob, and Carol
INSERT INTO todos (id, parent_id, list_id, user_id, title, description, completed, created_at, updated_at, complete_before, completed_at) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', NULL, 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Finish project report', 'Complete the final report for the XYZ project.', FALSE, '2024-01-01 10:15:00', '2024-01-01 10:15:00', '2024-01-05 17:00:00', NULL),
('fd17b30f-96c0-468a-99fe-c3745055d085', NULL, 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Finish project poc', NULL, FALSE, '2024-01-01 10:15:00', '2024-01-01 10:15:00', '2024-01-05 17:00:00', NULL),
('b2c3d4e5-f678-90ab-cdef-1234567890ab', NULL, 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Prepare presentation slides', NULL, TRUE, '2024-01-01 10:20:00', '2024-01-02 09:00:00', '2024-01-03 09:00:00', '2025-01-02 09:00:00'),
('9d63b9c4-78aa-4a46-b797-a12fe8fe4906', NULL, 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Prepare presentation speech', NULL, FALSE, '2024-01-01 10:20:00', '2024-01-02 09:00:00', '2024-01-03 09:00:00', '2025-01-02 09:00:00'),
('321f19e9-4eaf-44ff-87aa-7cace42d3c44', NULL, '4f3f5366-2a49-4bd0-8501-807416d49b56', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Buy groceries', 'Milk, eggs, bread, and fruits.', FALSE, '2024-01-01 10:25:00', '2024-01-01 10:25:00', NULL, NULL),
('c3d4e5f6-7890-abcd-ef12-34567890abcd', NULL, '4f3f5366-2a49-4bd0-8501-807416d49b56', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Fix bike', NULL, FALSE, '2024-01-01 10:25:00', '2024-01-01 10:25:00', NULL, NULL),
('d4e5f678-90ab-cdef-1234-567890abcdef', NULL, '0f137b9c-26fc-4c8e-924b-96edeb7d0500', '87b738aa-4fef-4334-b0d2-099614062ffd', 'Order new laptop', 'Research and order a new laptop for work.', FALSE, '2024-01-02 11:15:00', '2024-01-02 11:15:00', '2024-01-10 18:00:00', NULL),
('e5f67890-abcd-ef12-3456-7890abcdef12', NULL, 'd372481e-3fe1-4128-aba3-dee7d59ec9ce', '1fe81494-9017-440a-b432-d09541d8d273', 'Read "The Great Gatsby"', 'Finish reading the novel for book club.', TRUE, '2024-01-03 12:15:00', '2024-01-04 14:00:00', '2024-01-07 20:00:00', '2025-01-04 14:00:00'),
('f67890ab-cdef-1234-5678-90abcdef1234', NULL, 'd372481e-3fe1-4128-aba3-dee7d59ec9ce', '1fe81494-9017-440a-b432-d09541d8d273', 'Research Go programming', 'Look into advanced Go programming techniques.', FALSE, '2024-01-03 12:20:00', '2024-01-03 12:20:00', NULL, NULL);

-- Insert some sub tasks for Alice's "Finish project report" todo
INSERT INTO todos (id, parent_id, list_id, user_id, title, description, completed, created_at, updated_at, complete_before, completed_at) VALUES
('subtask-1', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Gather data', 'Collect all necessary data for the report.', TRUE, '2024-01-01 10:30:00', '2024-01-02 09:00:00', '2024-01-03 09:00:00', '2025-01-02 09:00:00'),
('subtask-2', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Draft report', 'Write the initial draft of the report.', FALSE, '2024-01-01 10:30:00', '2024-01-01 10:30:00', '2024-01-04 17:00:00', NULL),
('subtask-3', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'f60dc02e-8e56-4b7b-b157-73d58816e458', '9f080696-7101-4ec1-aa4e-8f3f78048a9e', 'Review and edit', 'Review the draft and make necessary edits.', FALSE, '2024-01-01 10:30:00', '2024-01-01 10:30:00', '2024-01-05 12:00:00', NULL);