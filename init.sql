-- Create the search table
CREATE TABLE search_data (
    id SERIAL PRIMARY KEY,
    text VARCHAR(50) NOT NULL
);


-- Insert 1 million distinct rows
INSERT INTO search_data (text)
SELECT 
    'sample_text_' || i || '_' || 
    CASE (i % 10)
        WHEN 0 THEN 'apple'
        WHEN 1 THEN 'banana'
        WHEN 2 THEN 'cherry'
        WHEN 3 THEN 'date'
        WHEN 4 THEN 'elderberry'
        WHEN 5 THEN 'fig'
        WHEN 6 THEN 'grape'
        WHEN 7 THEN 'honeydew'
        WHEN 8 THEN 'kiwi'
        ELSE 'lemon'
    END ||
    '_' || (i % 100)
FROM generate_series(1, 500000) AS i;