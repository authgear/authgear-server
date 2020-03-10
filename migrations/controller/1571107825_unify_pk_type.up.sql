ALTER TABLE app ADD COLUMN _id text UNIQUE;
UPDATE app SET _id = id;

ALTER TABLE artifact ADD COLUMN _id text UNIQUE;
UPDATE artifact SET _id = id;

ALTER TABLE cloud_code ADD COLUMN _id text UNIQUE;
UPDATE cloud_code SET _id = id;

ALTER TABLE cloud_code_backend ADD COLUMN _id text UNIQUE;
UPDATE cloud_code_backend SET _id = id;

ALTER TABLE config ADD COLUMN _id text UNIQUE;
UPDATE config SET _id = id;

ALTER TABLE deployment ADD COLUMN _id text UNIQUE;
UPDATE deployment SET _id = id;

ALTER TABLE deployment_hook ADD COLUMN _id text UNIQUE;
UPDATE deployment_hook SET _id = id;

ALTER TABLE deployment_route ADD COLUMN _id text UNIQUE;
UPDATE deployment_route SET _id = id;

ALTER TABLE domain ADD COLUMN _id text UNIQUE;
UPDATE domain SET _id = id;

ALTER TABLE invitation ADD COLUMN _id text UNIQUE;
UPDATE invitation SET _id = id;

ALTER TABLE microservice ADD COLUMN _id text UNIQUE;
UPDATE microservice SET _id = id;

ALTER TABLE plan ADD COLUMN _id text UNIQUE;
UPDATE plan SET _id = id;

ALTER TABLE secret ADD COLUMN _id text UNIQUE;
UPDATE secret SET _id = id;


ALTER TABLE app DROP CONSTRAINT app_plan_id_fkey;
ALTER TABLE app ALTER COLUMN plan_id TYPE text;
ALTER TABLE app ADD CONSTRAINT app_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES plan(_id);

ALTER TABLE app DROP CONSTRAINT app_config_id_fkey;
ALTER TABLE app ALTER COLUMN config_id TYPE text;
ALTER TABLE app ADD CONSTRAINT app_config_id_fkey FOREIGN KEY (config_id) REFERENCES config(_id);

ALTER TABLE app DROP CONSTRAINT app_last_deployment_id_fkey;
ALTER TABLE app ALTER COLUMN last_deployment_id TYPE text;
ALTER TABLE app ADD CONSTRAINT app_last_deployment_id_fkey FOREIGN KEY (last_deployment_id) REFERENCES deployment(_id);

ALTER TABLE app_user DROP CONSTRAINT app_user_app_id_fkey;
ALTER TABLE app_user ALTER COLUMN app_id TYPE text;
ALTER TABLE app_user ADD CONSTRAINT app_user_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE artifact DROP CONSTRAINT artifact_app_id_fkey;
ALTER TABLE artifact ALTER COLUMN app_id TYPE text;
ALTER TABLE artifact ADD CONSTRAINT artifact_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE cloud_code DROP CONSTRAINT cloud_code_app_id_fkey1;
ALTER TABLE cloud_code ALTER COLUMN app_id TYPE text;
ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE cloud_code DROP CONSTRAINT cloud_code_artifact_id_fkey;
ALTER TABLE cloud_code ALTER COLUMN artifact_id TYPE text;
ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_artifact_id_fkey FOREIGN KEY (artifact_id) REFERENCES artifact(_id);

ALTER TABLE cloud_code DROP CONSTRAINT cloud_code_backend_id_fkey;
ALTER TABLE cloud_code ALTER COLUMN backend_id TYPE text;
ALTER TABLE cloud_code ADD CONSTRAINT cloud_code_backend_id_fkey FOREIGN KEY (backend_id) REFERENCES cloud_code_backend(_id);

ALTER TABLE cloud_code_secret DROP CONSTRAINT cloud_code_secret_cloud_code_id_fkey;
ALTER TABLE cloud_code_secret ALTER COLUMN cloud_code_id TYPE text;
ALTER TABLE cloud_code_secret ADD CONSTRAINT cloud_code_secret_cloud_code_id_fkey FOREIGN KEY (cloud_code_id) REFERENCES cloud_code(_id);

ALTER TABLE cloud_code_secret DROP CONSTRAINT cloud_code_secret_secret_id_fkey;
ALTER TABLE cloud_code_secret ALTER COLUMN secret_id TYPE text;
ALTER TABLE cloud_code_secret ADD CONSTRAINT cloud_code_secret_secret_id_fkey FOREIGN KEY (secret_id) REFERENCES secret(_id);

ALTER TABLE deployment DROP CONSTRAINT deployment_app_id_fkey;
ALTER TABLE deployment ALTER COLUMN app_id TYPE text;
ALTER TABLE deployment ADD CONSTRAINT deployment_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE deployment_cloud_code DROP CONSTRAINT deployment_cloud_code_deployment_id_fkey;
ALTER TABLE deployment_cloud_code ALTER COLUMN deployment_id TYPE text;
ALTER TABLE deployment_cloud_code ADD CONSTRAINT deployment_cloud_code_deployment_id_fkey FOREIGN KEY (deployment_id) REFERENCES deployment(_id);

ALTER TABLE deployment_cloud_code DROP CONSTRAINT deployment_cloud_code_cloud_code_id_fkey;
ALTER TABLE deployment_cloud_code ALTER COLUMN cloud_code_id TYPE text;
ALTER TABLE deployment_cloud_code ADD CONSTRAINT deployment_cloud_code_cloud_code_id_fkey FOREIGN KEY (cloud_code_id) REFERENCES cloud_code(_id);

ALTER TABLE deployment_hook DROP CONSTRAINT deployment_hook_app_id_fkey;
ALTER TABLE deployment_hook ALTER COLUMN app_id TYPE text;
ALTER TABLE deployment_hook ADD CONSTRAINT deployment_hook_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE deployment_microservice DROP CONSTRAINT deployment_microservice_deployment_id_fkey;
ALTER TABLE deployment_microservice ALTER COLUMN deployment_id TYPE text;
ALTER TABLE deployment_microservice ADD CONSTRAINT deployment_microservice_deployment_id_fkey FOREIGN KEY (deployment_id) REFERENCES deployment(_id);

ALTER TABLE deployment_microservice DROP CONSTRAINT deployment_microservice_microservice_id_fkey;
ALTER TABLE deployment_microservice ALTER COLUMN microservice_id TYPE text;
ALTER TABLE deployment_microservice ADD CONSTRAINT deployment_microservice_microservice_id_fkey FOREIGN KEY (microservice_id) REFERENCES microservice(_id);

ALTER TABLE deployment_route DROP CONSTRAINT cloud_code_app_id_fkey;
ALTER TABLE deployment_route ALTER COLUMN app_id TYPE text;
ALTER TABLE deployment_route ADD CONSTRAINT deployment_route_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE domain DROP CONSTRAINT domain_app_id_fkey;
ALTER TABLE domain ALTER COLUMN app_id TYPE text;
ALTER TABLE domain ADD CONSTRAINT domain_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE invitation DROP CONSTRAINT invitation_app_id_fkey;
ALTER TABLE invitation ALTER COLUMN app_id TYPE text;
ALTER TABLE invitation ADD CONSTRAINT invitation_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE microservice DROP CONSTRAINT microservice_app_id_fkey;
ALTER TABLE microservice ALTER COLUMN app_id TYPE text;
ALTER TABLE microservice ADD CONSTRAINT microservice_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);

ALTER TABLE microservice DROP CONSTRAINT microservice_artifact_id_fkey;
ALTER TABLE microservice ALTER COLUMN artifact_id TYPE text;
ALTER TABLE microservice ADD CONSTRAINT microservice_artifact_id_fkey FOREIGN KEY (artifact_id) REFERENCES artifact(_id);

ALTER TABLE microservice_secret DROP CONSTRAINT microservice_secret_microservice_id_fkey;
ALTER TABLE microservice_secret ALTER COLUMN microservice_id TYPE text;
ALTER TABLE microservice_secret ADD CONSTRAINT microservice_secret_microservice_id_fkey FOREIGN KEY (microservice_id) REFERENCES microservice(_id);

ALTER TABLE microservice_secret DROP CONSTRAINT microservice_secret_secret_id_fkey;
ALTER TABLE microservice_secret ALTER COLUMN secret_id TYPE text;
ALTER TABLE microservice_secret ADD CONSTRAINT microservice_secret_secret_id_fkey FOREIGN KEY (secret_id) REFERENCES secret(_id);

ALTER TABLE secret DROP CONSTRAINT secret_app_id_fkey;
ALTER TABLE secret ALTER COLUMN app_id TYPE text;
ALTER TABLE secret ADD CONSTRAINT secret_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(_id);


ALTER TABLE app DROP COLUMN id;
ALTER TABLE app RENAME COLUMN _id TO id;
ALTER TABLE app ADD PRIMARY KEY (id);

ALTER TABLE artifact DROP COLUMN id;
ALTER TABLE artifact RENAME COLUMN _id TO id;
ALTER TABLE artifact ADD PRIMARY KEY (id);

ALTER TABLE cloud_code DROP COLUMN id;
ALTER TABLE cloud_code RENAME COLUMN _id TO id;
ALTER TABLE cloud_code ADD PRIMARY KEY (id);

ALTER TABLE cloud_code_backend DROP COLUMN id;
ALTER TABLE cloud_code_backend RENAME COLUMN _id TO id;
ALTER TABLE cloud_code_backend ADD PRIMARY KEY (id);

ALTER TABLE config DROP COLUMN id;
ALTER TABLE config RENAME COLUMN _id TO id;
ALTER TABLE config ADD PRIMARY KEY (id);

ALTER TABLE deployment DROP COLUMN id;
ALTER TABLE deployment RENAME COLUMN _id TO id;
ALTER TABLE deployment ADD PRIMARY KEY (id);

ALTER TABLE deployment_hook DROP COLUMN id;
ALTER TABLE deployment_hook RENAME COLUMN _id TO id;
ALTER TABLE deployment_hook ADD PRIMARY KEY (id);

ALTER TABLE deployment_route DROP COLUMN id;
ALTER TABLE deployment_route RENAME COLUMN _id TO id;
ALTER TABLE deployment_route ADD PRIMARY KEY (id);

ALTER TABLE domain DROP COLUMN id;
ALTER TABLE domain RENAME COLUMN _id TO id;
ALTER TABLE domain ADD PRIMARY KEY (id);

ALTER TABLE invitation DROP COLUMN id;
ALTER TABLE invitation RENAME COLUMN _id TO id;
ALTER TABLE invitation ADD PRIMARY KEY (id);

ALTER TABLE microservice DROP COLUMN id;
ALTER TABLE microservice RENAME COLUMN _id TO id;
ALTER TABLE microservice ADD PRIMARY KEY (id);

ALTER TABLE plan DROP COLUMN id;
ALTER TABLE plan RENAME COLUMN _id TO id;
ALTER TABLE plan ADD PRIMARY KEY (id);

ALTER TABLE secret DROP COLUMN id;
ALTER TABLE secret RENAME COLUMN _id TO id;
ALTER TABLE secret ADD PRIMARY KEY (id);
