DO $$
    BEGIN
        -- Check if the table exists
        IF EXISTS (SELECT FROM pg_catalog.pg_tables
                   WHERE schemaname = 'public' AND tablename  = 'users') THEN
            -- If the table exists, delete all rows from it
            EXECUTE 'DELETE FROM public.users';
        END IF;
    END$$;
