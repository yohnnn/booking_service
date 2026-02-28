-- +goose Up
-- +goose StatementBegin
INSERT INTO concerts (name, place, date, price, total_seats, available_seats) VALUES
    ('9mice', 'VK Stadium', '2026-05-17 20:00:00+03', 3000.00, 6500, 6500),
    ('Тёмный Принц', 'Base', '2026-06-08 22:00:00+03', 2500.00, 3500, 3500),
    ('fakemink', 'VK Stadium', '2026-04-25 19:30:00+03', 2000.00, 6500, 6500),
    ('9mice x kai angel', 'ВТБ Арена', '2026-07-12 20:00:00+03', 3500.00, 30500, 30500);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM concerts WHERE name IN (
    '9mice',
    'Тёмный Принц',
    'fakemink',
    '9mice x kai angel'
);
-- +goose StatementEnd
