DO LANGUAGE plpgsql
$$
DECLARE 
    fw VARCHAR;
BEGIN
    SELECT FROM information_schema.tables 
    INTO fw
    WHERE  table_schema = 'public' AND  table_name = 'flyway_schema_history';

    IF not found THEN 

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: live_reports; Type: TABLE; Schema: public; Owner: vulcan_reportgen
--

CREATE TABLE public.live_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email_subject text DEFAULT ''::text NOT NULL,
    email_body text DEFAULT ''::text NOT NULL,
    team_id text DEFAULT ''::text NOT NULL,
    date_to text DEFAULT ''::text NOT NULL,
    date_from text DEFAULT ''::text NOT NULL,
    delivered_to text DEFAULT ''::text NOT NULL,
    update_status_at timestamp with time zone DEFAULT now() NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.live_reports OWNER TO vulcan_reportgen;

--
-- Name: scan_reports; Type: TABLE; Schema: public; Owner: vulcan_reportgen
--

CREATE TABLE public.scan_reports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    scan_id text NOT NULL,
    report text DEFAULT ''::text NOT NULL,
    report_json text DEFAULT ''::text NOT NULL,
    email_subject text DEFAULT ''::text NOT NULL,
    email_body text DEFAULT ''::text NOT NULL,
    delivered_to text DEFAULT ''::text NOT NULL,
    update_status_at timestamp with time zone DEFAULT now() NOT NULL,
    status text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    program_name text NOT NULL,
    risk integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.scan_reports OWNER TO vulcan_reportgen;

--
-- Name: live_reports live_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: vulcan_reportgen
--

ALTER TABLE ONLY public.live_reports
    ADD CONSTRAINT live_reports_pkey PRIMARY KEY (id);


--
-- Name: live_reports live_reports_team_id_date_from_date_to_key; Type: CONSTRAINT; Schema: public; Owner: vulcan_reportgen
--

ALTER TABLE ONLY public.live_reports
    ADD CONSTRAINT live_reports_team_id_date_from_date_to_key UNIQUE (team_id, date_from, date_to);


--
-- Name: scan_reports scan_reports_pkey; Type: CONSTRAINT; Schema: public; Owner: vulcan_reportgen
--

ALTER TABLE ONLY public.scan_reports
    ADD CONSTRAINT scan_reports_pkey PRIMARY KEY (id);


--
-- Name: scan_reports scan_reports_scan_id_key; Type: CONSTRAINT; Schema: public; Owner: vulcan_reportgen
--

ALTER TABLE ONLY public.scan_reports
    ADD CONSTRAINT scan_reports_scan_id_key UNIQUE (scan_id);

    END IF;

END;
$$;
