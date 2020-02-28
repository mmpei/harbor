create table p2p_preheat_policy (
 id SERIAL NOT NULL,
 name varchar(256),
 project_id int NOT NULL,
 target_ids varchar(512),
 enabled boolean NOT NULL DEFAULT true,
 description text,
 filters varchar(1024),
 deleted boolean DEFAULT false NOT NULL,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

CREATE TRIGGER p2p_preheat_policy_update_time_at_modtime BEFORE UPDATE ON p2p_preheat_policy FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();

create table p2p_target (
 id SERIAL NOT NULL,
 name varchar(64),
 url varchar(64),
 username varchar(255),
 password varchar(128),
 /*
 type indicates the type of target p2p endpoint,
 0 means it's a kraken instance,
 1 means it's a dragonfly registry
 */
 type SMALLINT NOT NULL DEFAULT 0,
 insecure boolean NOT NULL DEFAULT false,
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

create table p2p_preheat_job (
 id SERIAL NOT NULL,
 status varchar(64) NOT NULL,
 policy_id int NOT NULL,
 repository varchar(256) NOT NULL,
 tag        varchar(256),
 job_uuid varchar(64),
 creation_time timestamp default CURRENT_TIMESTAMP,
 update_time timestamp default CURRENT_TIMESTAMP,
 PRIMARY KEY (id)
 );

CREATE INDEX p2p_preheat_job_policy ON p2p_preheat_job (policy_id);
CREATE INDEX p2p_preheat_job_poid_uptime ON p2p_preheat_job (policy_id, update_time);
CREATE INDEX p2p_preheat_job_poid_status ON p2p_preheat_job (policy_id, status);

CREATE TRIGGER p2p_preheat_job_update_time_at_modtime BEFORE UPDATE ON p2p_preheat_job FOR EACH ROW EXECUTE PROCEDURE update_update_time_at_column();
