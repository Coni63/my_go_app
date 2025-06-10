-- Liste des utilisateurs et schémas à créer
DO
$$
BEGIN
  -- Utilisateur et schéma pour auth_service
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'auth_service_admin') THEN
    CREATE USER auth_service_admin WITH PASSWORD 'auth_service_pass';
  END IF;
  EXECUTE 'CREATE SCHEMA IF NOT EXISTS auth_service AUTHORIZATION auth_service_admin';

  -- Utilisateur et schéma pour user_service
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'user_service_admin') THEN
    CREATE USER user_service_admin WITH PASSWORD 'user_service_pass';
  END IF;
  EXECUTE 'CREATE SCHEMA IF NOT EXISTS user_service AUTHORIZATION user_service_admin';

END
$$;