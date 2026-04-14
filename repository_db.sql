--
-- PostgreSQL database dump
--

\restrict DKCye47nOgfNvn43YL427Xe9wpOI4p13gqDV1v6ASgPYjFBGRTqf5HmaoLZLpEl

-- Dumped from database version 18.3
-- Dumped by pg_dump version 18.3

-- Started on 2026-04-14 11:11:27

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 222 (class 1259 OID 16440)
-- Name: access_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.access_requests (
    id uuid NOT NULL,
    document_id uuid NOT NULL,
    file_id uuid,
    nama character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    ktm_path character varying(500) DEFAULT ''::character varying,
    status character varying(50) DEFAULT 'pending'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    access_token character varying(100) DEFAULT ''::character varying
);


ALTER TABLE public.access_requests OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16454)
-- Name: document_files; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.document_files (
    id uuid NOT NULL,
    document_id uuid NOT NULL,
    file_name character varying(500) NOT NULL,
    file_path character varying(500) NOT NULL,
    file_size bigint DEFAULT 0,
    file_order integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_locked boolean DEFAULT false
);


ALTER TABLE public.document_files OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16467)
-- Name: document_views; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.document_views (
    id uuid NOT NULL,
    document_id uuid NOT NULL,
    ip_address character varying(100) DEFAULT ''::character varying,
    viewed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.document_views OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16474)
-- Name: documents; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.documents (
    id uuid NOT NULL,
    judul text NOT NULL,
    penulis text NOT NULL,
    jenis_file character varying(20) NOT NULL,
    file_path text NOT NULL,
    status character varying(10) NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    fakultas_id uuid,
    prodi_id uuid,
    dosen_pembimbing text DEFAULT ''::text,
    view_count integer DEFAULT 0,
    abstrak text DEFAULT ''::text,
    dosen_pembimbing_2 text DEFAULT ''::text,
    kata_kunci text DEFAULT ''::text,
    tahun integer DEFAULT 0
);


ALTER TABLE public.documents OWNER TO postgres;

--
-- TOC entry 226 (class 1259 OID 16492)
-- Name: email_otps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.email_otps (
    id uuid NOT NULL,
    email character varying(255) NOT NULL,
    otp_code character varying(10) NOT NULL,
    document_id uuid,
    is_verified boolean DEFAULT false,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.email_otps OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 16407)
-- Name: fakultas; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.fakultas (
    id uuid NOT NULL,
    nama character varying(255) NOT NULL,
    kode character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.fakultas OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16419)
-- Name: prodi; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.prodi (
    id uuid NOT NULL,
    nama character varying(255) NOT NULL,
    kode character varying(50) NOT NULL,
    fakultas_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.prodi OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16502)
-- Name: site_settings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.site_settings (
    key character varying(100) NOT NULL,
    value text DEFAULT ''::text,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.site_settings OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16599)
-- Name: site_visits; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.site_visits (
    id uuid NOT NULL,
    ip_address character varying(100) DEFAULT ''::character varying,
    visited_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.site_visits OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16510)
-- Name: student_registrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.student_registrations (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    ktm_path character varying(500) DEFAULT ''::character varying,
    status character varying(50) DEFAULT 'pending'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.student_registrations OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16389)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    password character varying(255) NOT NULL,
    role character varying(50) DEFAULT 'user'::character varying,
    created_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT users_role_check CHECK (((role)::text = ANY ((ARRAY['admin'::character varying, 'user'::character varying, 'mahasiswa'::character varying])::text[])))
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 5129 (class 0 OID 16440)
-- Dependencies: 222
-- Data for Name: access_requests; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.access_requests (id, document_id, file_id, nama, email, ktm_path, status, created_at, updated_at, access_token) FROM stdin;
c71dcd20-63fb-489a-a6a8-79daa2a4a7cb	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	\N	Orenji Snack	rendhirichardo7@gmail.com	1glaiFwViC18bPqseNRGaAdCTIvtgoPT2	approved	2026-04-08 13:24:42.699079+07	2026-04-08 13:24:57.13407+07	5464674359a2276f
\.


--
-- TOC entry 5130 (class 0 OID 16454)
-- Dependencies: 223
-- Data for Name: document_files; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.document_files (id, document_id, file_name, file_path, file_size, file_order, created_at, is_locked) FROM stdin;
c6484f92-f3a2-4dcb-a31d-1fff87459bfd	239f841d-a823-47ad-82d8-dd452d08fd1a	BAB I new.pdf	16A4cZBdLMYSu6HJgXL9f46M2xmLdQyxM	581359	0	2026-04-08 15:07:37.664724+07	f
c0a7b24d-c4c7-4b64-b64f-fe255c40a63e	239f841d-a823-47ad-82d8-dd452d08fd1a	BAB II new.pdf	147nKdWbHQ3VSBLPXErN2QSUSqFj28_J8	528073	1	2026-04-08 15:07:37.665181+07	f
be22e561-6abb-4816-a26d-605feac01b53	239f841d-a823-47ad-82d8-dd452d08fd1a	BAB III new.pdf	16LrwIK6WhzWDMXiB5qRvlru2PbXrc8bk	4062461	2	2026-04-08 15:07:37.665503+07	f
f60bf18c-5886-4940-8f0f-fb9fed7d2a72	239f841d-a823-47ad-82d8-dd452d08fd1a	BAB IV new.pdf	1Dnu23xUVAFM9Xffj5RNC6UKYCs4oSag7	5437624	3	2026-04-08 15:07:37.665775+07	f
4ed782cd-697f-494f-a151-039406a5f55b	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB I new.pdf	1Mqg_VwGJzET8s9xHeCFeiNB3Aw58NaXp	581359	0	2026-04-08 13:06:23.825278+07	f
107e645f-a637-4f34-8d95-9ac835ba32e8	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB II new.pdf	1pAS72hfpoz-CfpoW5TDQXE54T6ZXEQJj	528073	1	2026-04-08 13:06:23.827923+07	f
fdb4c11e-5256-4c37-9df7-306bb7d072f8	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB III new.pdf	19KcFObW6vYQ8qWnE7WH5bB-Pka9ADfW1	4062461	2	2026-04-08 13:06:23.829342+07	t
b82091fe-002c-45b3-80d4-db56277f0841	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB IV new.pdf	1GHl62yNHZWB9WGbyR0PvxDAglBDz9N9I	5437624	3	2026-04-08 13:06:23.83081+07	t
11712186-55d9-4fb1-98e0-848dcf80f7f6	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB V.pdf	13T9z-OGIvYUGP-iRCcpwajW_UT0z1Ldv	6115093	4	2026-04-08 13:06:23.832145+07	t
60b51029-66a0-44a5-8bef-72f1c030d502	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	BAB VI.pdf	1pazoerOW1_HIKAD4Cll2cUW2H0_Juf1x	271466	5	2026-04-08 13:06:23.833623+07	t
b88b23dc-c40d-482e-aee4-136bcdc53e0a	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	cover.pdf	1zf3nXMv7HoA5ee-KmmLVNGBIYOaePCkD	204835	6	2026-04-08 13:06:23.834986+07	t
ab5bd9fd-e81c-4d04-9a15-c65d43324356	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	Daftar Pustaka.pdf	1lS6AtiH6gpaZ9r5nBgvUQWIoil0_6ofj	285837	7	2026-04-08 13:06:23.838247+07	t
31a625cd-a54b-4567-94ee-8e004924bfae	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	KATA PENGANTAR dan DAFTAR ISI SKRIPSI.pdf	1iWYIze2uJG005QWw_yKSvBaLzNqmL74A	954489	8	2026-04-08 13:06:23.839535+07	t
855f5af9-2c3f-4da8-b9df-a43929397e14	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	Lampiran Komprehensif new.pdf	1H8HPQmjoc578U_St0SZVNzl1m7LCnVbA	832651	9	2026-04-08 13:06:23.84069+07	t
0dde3bbf-1f84-4ad6-a5c9-85a1f5019dd7	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	TINJAUAN PUSTAKA.pdf	1kfnCnBTFFHuTlVHSx6irLmDTdBRT4ZvA	215999	10	2026-04-08 13:06:23.841744+07	t
e87a11ba-c2b0-4c6e-a7f9-1e192f413d60	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB I new.pdf	1rZmm2xGAFZAy2a3eVqyX7ivOD5u4BujW	581359	0	2026-04-08 13:11:59.91629+07	f
bfcd10b9-5a06-4d80-85c8-76a7bf557dd6	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB II new.pdf	1Wp3IjiV06FYY5CWHPSQfooy_r3w_Zt8D	528073	1	2026-04-08 13:11:59.917112+07	f
7ae6977e-362a-48f7-8667-26737f327e5f	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB III new.pdf	1GkRvu3weHrqFJa3jfOF5QZyKJRSMsf69	4062461	2	2026-04-08 13:11:59.917449+07	t
1686b2e8-7d9d-4800-a2e6-de7ea9842b07	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB IV new.pdf	1JrWZIqQ6MZ2vZwsqbOjNYuuSDmwluW-j	5437624	3	2026-04-08 13:11:59.917764+07	t
112bb3f0-eba3-429d-82ba-86b428f77e08	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB V.pdf	1VbAVwhvdftkU0c2-Xa6On6U2HkG1Q6K3	6115093	4	2026-04-08 13:11:59.918073+07	t
ddc5ab97-62d6-4e3b-ab00-844d5b484e5b	80ab58bf-36b5-4ebe-bc94-f10375f84a4a	BAB VI.pdf	1A2Lqqv0PZard43nTPHT0P7R0np3WV_r9	271466	5	2026-04-08 13:11:59.918363+07	t
7bbb3da3-52f1-4ed6-a92c-921e2f2301ff	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB I new.pdf	13VRs6Dk14JH0evxL3vLbQbtV7a9CQ-KI	581359	0	2026-04-08 13:17:09.950684+07	f
e148c514-4251-4803-89c0-4c490573b178	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB II new.pdf	1iifpWq8DyQg1iW4iSzv9e9jzxgTI3UxX	528073	1	2026-04-08 13:17:09.951535+07	f
e9f40464-90d1-4c67-9ef9-1998d6204ec1	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB III new.pdf	1BTKWlgpWT2hG6bEJtQ-51ch9z1njf-3F	4062461	2	2026-04-08 13:17:09.9519+07	f
f1f99c15-f367-451b-a64f-9e91e578ae8a	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB IV new.pdf	1mwzQsssxnrC7d8sa3VXiTIIAva010CFk	5437624	3	2026-04-08 13:17:09.952259+07	f
0eb56a09-9c89-44c5-9eea-005a1e689624	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB V.pdf	1ye6jjRGDn05kPWEKWCN8qnoglhdSQG0c	6115093	4	2026-04-08 13:17:09.952595+07	f
dc407a88-f2c0-4f1b-bb7a-14d58b288aee	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	BAB VI.pdf	1iRG-zWaTPAGFeb4AZ4aE0PVHtVsH3S7u	271466	5	2026-04-08 13:17:09.952944+07	f
4f1feb1f-b31a-4c6e-a707-290dd7869695	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	cover.pdf	11Rl-UK100b5CB_lstOnrdNmWAcwfpJxp	204835	6	2026-04-08 13:17:09.953281+07	f
5211171c-2ea2-4393-a2b8-038734b1f35a	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	Daftar Pustaka.pdf	1QhYEDrFJEUhpBsmF0rsn__MYON3IAJGg	285837	7	2026-04-08 13:17:09.953526+07	f
cddd8c84-92a9-42fa-a763-418c5550bbc3	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	KATA PENGANTAR dan DAFTAR ISI SKRIPSI.pdf	1UOA9fn8ujQRJag-MAbW1XcavbMB0xs37	954489	8	2026-04-08 13:17:09.953783+07	f
735a9036-74e3-408d-aee0-bb6a18337e25	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	Lampiran Komprehensif new.pdf	1RbwAhT5cESzP3Bp9os8li9amPOfme5HB	832651	9	2026-04-08 13:17:09.954035+07	f
17f021b7-e22c-4696-9f11-418b346f5352	2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	TINJAUAN PUSTAKA.pdf	1jxlWkWQJf2RH9UgZPsQlwKnJtkwrRIik	215999	10	2026-04-08 13:17:09.954327+07	f
13f40d9f-2f89-48f5-b1ad-26facbaeb26e	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB I new.pdf	uploads/736b7a95-1cab-428d-ab27-f3cb252b882d.pdf	581359	0	2026-04-08 15:10:41.734952+07	f
b8be9f71-c91d-43ae-8348-4fac0c99b49b	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB II new.pdf	uploads/1486e2be-ba86-45b0-b433-0ab3b93053ea.pdf	528073	1	2026-04-08 15:10:41.735672+07	f
cb3c7cbb-5baf-4f44-b4bc-d8b0ea34ef71	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB III new.pdf	uploads/1a66c14c-ffad-4b95-bbd3-66552bc1e129.pdf	4062461	2	2026-04-08 15:10:41.736133+07	f
e375749b-6aee-4c93-ab90-20569b0a767f	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB IV new.pdf	uploads/78b666b5-701e-49ce-9d5f-3f1a6537223a.pdf	5437624	3	2026-04-08 15:10:41.736576+07	f
ecef3295-d650-4b91-a12a-2fd7363b193f	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB V.pdf	uploads/fa86056b-c7c3-4516-bdf2-61051cb76c6d.pdf	6115093	4	2026-04-08 15:10:41.737034+07	f
9f4f594b-5cf3-43e8-a020-1de3380eca12	0971fa8a-4bcf-4440-a5d5-2ba71646c762	BAB VI.pdf	uploads/9a5b6d6a-eea1-445a-8d9d-0963130c6c23.pdf	271466	5	2026-04-08 15:10:41.737424+07	f
0d0509b8-8f15-4dfd-93a3-294330df669e	0971fa8a-4bcf-4440-a5d5-2ba71646c762	cover.pdf	uploads/91aa7120-53c3-4c1c-915f-0b23484c2e24.pdf	204835	6	2026-04-08 15:10:41.737814+07	f
97603921-d868-4a76-8e32-d208d1126940	0971fa8a-4bcf-4440-a5d5-2ba71646c762	Daftar Pustaka.pdf	uploads/b419fbde-eefe-4bb1-a174-403c04039d92.pdf	285837	7	2026-04-08 15:10:41.738181+07	f
36f6cbe2-553f-4922-bf91-a370504fa4ce	0971fa8a-4bcf-4440-a5d5-2ba71646c762	KATA PENGANTAR dan DAFTAR ISI SKRIPSI.pdf	uploads/7dbf0e0d-63bf-4f8c-af8f-8bd29f19f6b0.pdf	954489	8	2026-04-08 15:10:41.73849+07	f
4ea4dea2-483f-43ea-bd5c-0cc75978680a	0971fa8a-4bcf-4440-a5d5-2ba71646c762	Lampiran Komprehensif new.pdf	uploads/4badabe0-25fe-4fb5-9a01-6dbb12eaa175.pdf	832651	9	2026-04-08 15:10:41.738847+07	f
eeaf606e-1790-433c-9729-37ccec05a935	0971fa8a-4bcf-4440-a5d5-2ba71646c762	TINJAUAN PUSTAKA.pdf	uploads/82072499-da23-4bde-aa58-83b6881c9050.pdf	215999	10	2026-04-08 15:10:41.739163+07	f
\.


--
-- TOC entry 5131 (class 0 OID 16467)
-- Dependencies: 224
-- Data for Name: document_views; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.document_views (id, document_id, ip_address, viewed_at) FROM stdin;
\.


--
-- TOC entry 5132 (class 0 OID 16474)
-- Dependencies: 225
-- Data for Name: documents; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.documents (id, judul, penulis, jenis_file, file_path, status, created_at, fakultas_id, prodi_id, dosen_pembimbing, view_count, abstrak, dosen_pembimbing_2, kata_kunci, tahun) FROM stdin;
d328f182-0fbd-436f-a3e1-4a9354ecd6c8	SISTEM INFORMASI E-MARKETPLACE BENDA SENI PADA KELOMPOK PEDAGANG SIGERTENGAH	Caca Arif Herdian	skripsi	1Mqg_VwGJzET8s9xHeCFeiNB3Aw58NaXp	publish	2026-04-08 13:06:23.820032	e98bb137-2acd-4d39-9868-efc11eb7f2b7	c4a7f30e-3e1d-4d42-9f91-ee37942659af	Rakhmayudhi, M.Kom	0	Sampai saat ini banyak e-marketplace yang telah bermunculan di indonesia, namun tidak ada yang dikhususkan untuk komunitas atau kelompok dengan hobi yang sama di bidang seni. Kelompok pedagang sigertengah merupakan organisasi yang bergerak di bidang seni dengan memproduksi dan menjual benda-benda seni. Permasalahan yang sering terjadi dalam proses penjualan benda seni hasil produksinya sering mengalami ketidakpastian dari pembeli yaitu tidak pastinya ketersediaan uang yang dimiliki oleh pembeli serta pedagang yang tidak ingin menampilkan nominal harga dari benda seni yang ditawarkan, hal ini menyebabkan rendahnya kepastian pembelian. Oleh karena itu, perlu dilakukan perubahan konsep proses penjualan. Dalam menyelesaikan masalah tersebut penulis membangun sistem informasi e-marketplace dan merubah konsep tata cara pembayaran dengan menerapkan cara pembayaran escrow atau rekening bersama yang disimpan oleh pihak ketiga dalam hal ini pengelola marketplace . Sistem informasi ini dibangun menggunakan metode pengembangan sistem Unified Process , bahasa pemrograman PHP dengan Framework Codeigniter , database MySQL serta bahasa pemodelan UML .	Bambang Tjahjo Utomo, MT	Sistem Informasi, E-Marketplace , Escrow , Unified Process , PHP ,	2018
80ab58bf-36b5-4ebe-bc94-f10375f84a4a	ajsdnajkdnlasjdn	asep	skripsi	1rZmm2xGAFZAy2a3eVqyX7ivOD5u4BujW	publish	2026-04-08 13:11:59.907376	e98bb137-2acd-4d39-9868-efc11eb7f2b7	c4a7f30e-3e1d-4d42-9f91-ee37942659af	adxzcxz	0	mn amnbadbsakjdbsalidubasdiasbdjsadnasd	fgerteds	adasd	2018
2ff1fc3c-1a2b-44b5-badc-2deffe5524b0	iashdasihdksajdnsajd	aslkdaskdj	skripsi	13VRs6Dk14JH0evxL3vLbQbtV7a9CQ-KI	publish	2026-04-08 13:17:09.941538	e98bb137-2acd-4d39-9868-efc11eb7f2b7	c4a7f30e-3e1d-4d42-9f91-ee37942659af	Rakhmayudhi, M.Kom	0	Sampai saat ini banyak e-marketplace yang telah bermunculan di indonesia, namun tidak ada yang dikhususkan untuk komunitas atau kelompok dengan hobi yang sama di bidang seni. Kelompok pedagang sigertengah merupakan organisasi yang bergerak di bidang seni dengan memproduksi dan menjual benda-benda seni. Permasalahan yang sering terjadi dalam proses penjualan benda seni hasil produksinya sering mengalami ketidakpastian dari pembeli yaitu tidak pastinya ketersediaan uang yang dimiliki oleh pembeli serta pedagang yang tidak ingin menampilkan nominal harga dari benda seni yang ditawarkan, hal ini menyebabkan rendahnya kepastian pembelian. Oleh karena itu, perlu dilakukan perubahan konsep proses penjualan. Dalam menyelesaikan masalah tersebut penulis membangun sistem informasi e-marketplace dan merubah konsep tata cara pembayaran dengan menerapkan cara pembayaran escrow atau rekening bersama yang disimpan oleh pihak ketiga dalam hal ini pengelola marketplace . Sistem informasi ini dibangun menggunakan metode pengembangan sistem Unified Process , bahasa pemrograman PHP dengan Framework Codeigniter , database MySQL serta bahasa pemodelan UML .	Bambang Tjahjo Utomo, MT	ajbsdkasbdmnasd,msadn	2019
239f841d-a823-47ad-82d8-dd452d08fd1a	dfvghbnjk	rdftgyhujk	skripsi	16A4cZBdLMYSu6HJgXL9f46M2xmLdQyxM	publish	2026-04-08 15:07:37.663335	17b45ddf-4ac5-43ce-9be2-13aefe876c4a	7e8e4c44-007e-4a6b-9ce8-2c7db427562e	dfg	0	ijohu8gy7tf65	ikloj	wedrftgyh	2021
0971fa8a-4bcf-4440-a5d5-2ba71646c762	SISTEM INFORMASI E-MARKETPLACE BENDA SENI PADA KELOMPOK PEDAGANG SIGERTENGAH 	Caca Arif Herdian	skripsi	uploads/736b7a95-1cab-428d-ab27-f3cb252b882d.pdf	publish	2026-04-08 15:10:41.726187	e98bb137-2acd-4d39-9868-efc11eb7f2b7	c4a7f30e-3e1d-4d42-9f91-ee37942659af	Rakhmayudhi, M.Kom	0	Oleh : Caca Arif Herdian Sampai saat ini banyak e-marketplace yang telah bermunculan di indonesia, namun tidak ada yang dikhususkan untuk komunitas atau kelompok dengan hobi yang sama di bidang seni. Kelompok pedagang sigertengah merupakan organisasi yang bergerak di bidang seni dengan memproduksi dan menjual benda-benda seni. Permasalahan yang sering terjadi dalam proses penjualan benda seni hasil produksinya sering mengalami ketidakpastian dari pembeli yaitu tidak pastinya ketersediaan uang yang dimiliki oleh pembeli serta pedagang yang tidak ingin menampilkan nominal harga dari benda seni yang ditawarkan, hal ini menyebabkan rendahnya kepastian pembelian. Oleh karena itu, perlu dilakukan perubahan konsep proses penjualan. Dalam menyelesaikan masalah tersebut penulis membangun sistem informasi e-marketplace dan merubah konsep tata cara pembayaran dengan menerapkan cara pembayaran escrow atau rekening bersama yang disimpan oleh pihak ketiga dalam hal ini pengelola marketplace . Sistem informasi ini dibangun menggunakan metode pengembangan sistem Unified Process , bahasa pemrograman PHP dengan Framework Codeigniter , database MySQL serta bahasa pemodelan UML .	Bambang Tjahjo Utomo, MT	Sistem Informasi, E-Marketplace , Escrow , Unified Process , PHP ,	2024
\.


--
-- TOC entry 5133 (class 0 OID 16492)
-- Dependencies: 226
-- Data for Name: email_otps; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.email_otps (id, email, otp_code, document_id, is_verified, expires_at, created_at) FROM stdin;
947352cc-15ea-4c24-9916-ee723f17d4d1	rendhirichardo1@unsub.ac.id	383810	5ca7b8e0-d01e-41a3-bf6f-693a48e443d8	t	2026-03-02 13:30:22.975901+07	2026-03-02 13:25:22.977132+07
9e3f87fb-53ea-4b78-8971-b6272665ee51	perpustakaan@unsub.ac.id	619354	3a29d68e-6681-4e3c-a68c-c0a30d0e4009	t	2026-03-02 13:54:54.239685+07	2026-03-02 13:49:54.240487+07
f9d8f864-ff2d-45a7-b909-e20edfe5722a	rendhirichardo1@unsub.ac.id	767096	55723384-649e-4a41-b0d5-6e91ea0c047d	t	2026-03-03 15:26:07.550498+07	2026-03-03 15:21:07.553408+07
41d34c82-901e-4cb4-91f3-cd556b1419f4	rendhirichardo7@gmail.com	738189	3a29d68e-6681-4e3c-a68c-c0a30d0e4009	t	2026-03-04 14:34:44.264385+07	2026-03-04 14:29:44.273737+07
9a8eb093-6414-4929-848a-15d7bf47e1a0	rendhirichardo7@gmial.com	978320	\N	f	2026-03-30 11:23:05.566495+07	2026-03-30 11:18:05.568731+07
ebe21a29-e8d2-4b3f-b8a4-0a9a5cc34370	aristarafelia@gmail.com	547270	fb28a05d-36c6-42f5-a353-6dfa8e2ae06e	t	2026-03-30 11:47:59.316088+07	2026-03-30 11:42:59.318412+07
6da936a7-930c-4847-95f3-f69f3f6d4bfa	rendhirichardo1@unsub.ac.id	109854	\N	t	2026-04-08 12:46:33.978002+07	2026-04-08 12:41:33.98734+07
da3f274d-ceb6-4c83-83d0-296672af984a	rendhirichardo7@gmail.com	449220	\N	t	2026-04-08 13:23:30.453063+07	2026-04-08 13:18:30.454338+07
8cce6428-fe10-4290-88e6-fb5890b7744e	rendhirichardo7@gmail.com	722882	d328f182-0fbd-436f-a3e1-4a9354ecd6c8	t	2026-04-08 13:29:22.187393+07	2026-04-08 13:24:22.188417+07
\.


--
-- TOC entry 5127 (class 0 OID 16407)
-- Dependencies: 220
-- Data for Name: fakultas; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.fakultas (id, nama, kode, created_at, updated_at) FROM stdin;
e98bb137-2acd-4d39-9868-efc11eb7f2b7	Fakultas Ilmu Komputer	FASILKOM	2026-02-11 22:51:03.205735+07	2026-02-11 22:51:03.205735+07
17b45ddf-4ac5-43ce-9be2-13aefe876c4a	Fakultas Ilmu Administrasi	FIA	2026-02-23 13:59:34.695478+07	2026-02-23 13:59:34.695478+07
50b6cba8-cd41-4e0b-b451-842bd4035ba5	Fakultas Keguruan dan Ilmu Pendidikan	FKIP	2026-02-23 13:59:47.185142+07	2026-02-23 13:59:47.185142+07
735f8d19-1f5c-4b01-a790-2b2055e76a5a	Fakultas Hukum	FHUM	2026-03-30 11:13:21.183108+07	2026-03-30 11:13:21.183108+07
\.


--
-- TOC entry 5128 (class 0 OID 16419)
-- Dependencies: 221
-- Data for Name: prodi; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.prodi (id, nama, kode, fakultas_id, created_at, updated_at) FROM stdin;
c4a7f30e-3e1d-4d42-9f91-ee37942659af	Sistem Informasi	SI	e98bb137-2acd-4d39-9868-efc11eb7f2b7	2026-02-11 22:51:19.781101+07	2026-02-11 22:51:19.781101+07
64af647f-f602-4153-9b29-97cf279def3d	Administrasi Publik	AP	17b45ddf-4ac5-43ce-9be2-13aefe876c4a	2026-02-23 14:00:16.664756+07	2026-02-23 14:00:16.664756+07
adf4021b-ed8c-4dec-99b9-2c8908e26e96	Administrasi Bisnis	AB	17b45ddf-4ac5-43ce-9be2-13aefe876c4a	2026-02-23 14:00:22.845339+07	2026-02-23 14:00:22.845339+07
7e8e4c44-007e-4a6b-9ce8-2c7db427562e	Administrasi Keuangan	AK	17b45ddf-4ac5-43ce-9be2-13aefe876c4a	2026-02-23 14:00:29.834203+07	2026-02-23 14:00:29.834203+07
5f937c86-6d44-44af-adc2-995ce37df130	Pendidikan Jasmani, Kesehatan, dan Rekreasi	PJKR	50b6cba8-cd41-4e0b-b451-842bd4035ba5	2026-02-23 14:01:27.060682+07	2026-02-23 14:01:27.060682+07
4ed24bbf-4a47-4866-a1a5-ccbda5883038	Pendidikan Matematika	PM	50b6cba8-cd41-4e0b-b451-842bd4035ba5	2026-02-23 14:01:38.630526+07	2026-02-23 14:01:38.630526+07
ebfa3df0-b157-4f29-99ba-3ee9c950b7ba	Pendidikan Bahasa Inggris	PBI	50b6cba8-cd41-4e0b-b451-842bd4035ba5	2026-02-23 14:01:51.323665+07	2026-02-23 14:01:51.323665+07
ea89e1fb-1700-4230-be0a-6f322dd0626a	Ilmu Hukum	IH	735f8d19-1f5c-4b01-a790-2b2055e76a5a	2026-03-30 11:13:57.244712+07	2026-03-30 11:13:57.244712+07
\.


--
-- TOC entry 5134 (class 0 OID 16502)
-- Dependencies: 227
-- Data for Name: site_settings; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.site_settings (key, value, updated_at) FROM stdin;
logo_url	/uploads/logo/site_logo_5cf0c5aa-9afd-42ce-948c-cb825a35d9ea.png	2026-02-15 16:04:01.520482+07
views_reset_v1	done	2026-03-03 15:32:57.016147+07
site_visits_dedup_v1	done	2026-03-30 11:55:53.692752+07
site_visits_reset_v1	done	2026-03-30 11:57:40.506586+07
views_dedup_v1	done	2026-03-30 12:24:04.665399+07
misi		2026-04-08 12:09:56.365137+07
phone	000000000	2026-04-08 12:09:56.37188+07
app_name	Repository Universitas Subang	2026-04-08 12:09:56.373266+07
address	Jalan RA Kartini KM 3, Desa/Kelurahan Dangdeur (sebelumnya tercatat Nyimplung), Kecamatan Subang	2026-04-08 12:09:56.377841+07
email	perpustakaan@unsub.ac.id	2026-04-08 12:09:56.378435+07
footer_text	2024 Universitas Subang	2026-04-08 12:09:56.378969+07
app_description	Repository Universitas Subang	2026-04-08 12:09:56.380287+07
about_text	Platform digital untuk menyimpan, mengelola, dan menyebarluaskan karya ilmiah civitas akademika Universitas subang.	2026-04-08 12:09:56.381583+07
visi		2026-04-08 12:09:56.382076+07
\.


--
-- TOC entry 5136 (class 0 OID 16599)
-- Dependencies: 229
-- Data for Name: site_visits; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.site_visits (id, ip_address, visited_at) FROM stdin;
61875102-5c6d-4697-bc4d-2ebfe8c74996	10.10.4.200	2026-03-30 11:57:53.871451+07
9bbf7125-f70e-4258-aeb0-e5ca1cc765df	::1	2026-04-08 12:00:42.46953+07
73ce5aae-3077-416e-991c-624a2974db80	192.168.17.173	2026-04-08 14:13:26.048079+07
\.


--
-- TOC entry 5135 (class 0 OID 16510)
-- Dependencies: 228
-- Data for Name: student_registrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.student_registrations (id, name, email, password, ktm_path, status, created_at, updated_at) FROM stdin;
a7435d06-0b3e-4b15-b6cb-a05c990aae42	KMB Miftahul Huda	rendhirichardo1@unsub.ac.id	$2a$10$1lXeby.Ch/0wADstu68ZIO0bJg6NMOrTYI8jaj/uq3VD8JAFMjgdu	1EMUlqkjAJ979mIkMDTT8Kqiwvbxp0xp7	approved	2026-04-08 12:42:08.165022+07	2026-04-08 12:42:33.749536+07
c1a6cbfe-e901-47d1-be2e-4941a65e6b09	Orenji Snack	rendhirichardo7@gmail.com	$2a$10$vtB5wws5DAzdcjHR2p1GOOVyNAjLdQGrvdQArZ8VkyUcXgKfaGlx6	1rrlavVr1Rd0AUWTzy9GHy9aDtgP8OPix	approved	2026-04-08 13:19:03.386128+07	2026-04-08 13:19:28.525417+07
\.


--
-- TOC entry 5126 (class 0 OID 16389)
-- Dependencies: 219
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, name, email, password, role, created_at, updated_at) FROM stdin;
b63bf4f3-cc3c-4716-b86d-48af560b672e	rendhi richardo	rendhirichardo2@gmail.com	$2a$10$c0fx7YCs308aW9XcJyJhEe.WhMfIBV/hXoCOFjg/FkVAVRLdSWiNS	admin	2026-03-02 13:08:14.026617+07	2026-03-02 13:08:14.026617+07
e6c92606-2165-4904-b47b-001d871bd3e3	Orenji Snack	rendhirichardo7@gmail.com	$2a$10$vtB5wws5DAzdcjHR2p1GOOVyNAjLdQGrvdQArZ8VkyUcXgKfaGlx6	mahasiswa	2026-04-08 13:19:28.515895+07	2026-04-08 13:19:28.515895+07
\.


--
-- TOC entry 4948 (class 2606 OID 16524)
-- Name: access_requests access_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.access_requests
    ADD CONSTRAINT access_requests_pkey PRIMARY KEY (id);


--
-- TOC entry 4953 (class 2606 OID 16526)
-- Name: document_files document_files_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.document_files
    ADD CONSTRAINT document_files_pkey PRIMARY KEY (id);


--
-- TOC entry 4956 (class 2606 OID 16528)
-- Name: document_views document_views_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.document_views
    ADD CONSTRAINT document_views_pkey PRIMARY KEY (id);


--
-- TOC entry 4959 (class 2606 OID 16530)
-- Name: documents documents_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);


--
-- TOC entry 4961 (class 2606 OID 16532)
-- Name: email_otps email_otps_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.email_otps
    ADD CONSTRAINT email_otps_pkey PRIMARY KEY (id);


--
-- TOC entry 4937 (class 2606 OID 16418)
-- Name: fakultas fakultas_kode_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fakultas
    ADD CONSTRAINT fakultas_kode_key UNIQUE (kode);


--
-- TOC entry 4939 (class 2606 OID 16416)
-- Name: fakultas fakultas_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fakultas
    ADD CONSTRAINT fakultas_pkey PRIMARY KEY (id);


--
-- TOC entry 4944 (class 2606 OID 16431)
-- Name: prodi prodi_kode_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prodi
    ADD CONSTRAINT prodi_kode_key UNIQUE (kode);


--
-- TOC entry 4946 (class 2606 OID 16429)
-- Name: prodi prodi_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prodi
    ADD CONSTRAINT prodi_pkey PRIMARY KEY (id);


--
-- TOC entry 4965 (class 2606 OID 16534)
-- Name: site_settings site_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.site_settings
    ADD CONSTRAINT site_settings_pkey PRIMARY KEY (key);


--
-- TOC entry 4971 (class 2606 OID 16606)
-- Name: site_visits site_visits_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.site_visits
    ADD CONSTRAINT site_visits_pkey PRIMARY KEY (id);


--
-- TOC entry 4969 (class 2606 OID 16536)
-- Name: student_registrations student_registrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.student_registrations
    ADD CONSTRAINT student_registrations_pkey PRIMARY KEY (id);


--
-- TOC entry 4933 (class 2606 OID 16405)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4935 (class 2606 OID 16403)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4949 (class 1259 OID 16537)
-- Name: idx_access_requests_doc_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_access_requests_doc_id ON public.access_requests USING btree (document_id);


--
-- TOC entry 4950 (class 1259 OID 16538)
-- Name: idx_access_requests_file_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_access_requests_file_id ON public.access_requests USING btree (file_id);


--
-- TOC entry 4951 (class 1259 OID 16539)
-- Name: idx_access_requests_token; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_access_requests_token ON public.access_requests USING btree (access_token);


--
-- TOC entry 4954 (class 1259 OID 16540)
-- Name: idx_document_files_doc_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_document_files_doc_id ON public.document_files USING btree (document_id);


--
-- TOC entry 4957 (class 1259 OID 16541)
-- Name: idx_document_views_doc_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_document_views_doc_id ON public.document_views USING btree (document_id);


--
-- TOC entry 4962 (class 1259 OID 16542)
-- Name: idx_email_otps_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_otps_email ON public.email_otps USING btree (email);


--
-- TOC entry 4963 (class 1259 OID 16543)
-- Name: idx_email_otps_expires; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_email_otps_expires ON public.email_otps USING btree (expires_at);


--
-- TOC entry 4940 (class 1259 OID 16437)
-- Name: idx_fakultas_kode; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_fakultas_kode ON public.fakultas USING btree (kode);


--
-- TOC entry 4941 (class 1259 OID 16439)
-- Name: idx_prodi_fakultas_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_prodi_fakultas_id ON public.prodi USING btree (fakultas_id);


--
-- TOC entry 4942 (class 1259 OID 16438)
-- Name: idx_prodi_kode; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_prodi_kode ON public.prodi USING btree (kode);


--
-- TOC entry 4966 (class 1259 OID 16544)
-- Name: idx_student_registrations_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_student_registrations_email ON public.student_registrations USING btree (email);


--
-- TOC entry 4967 (class 1259 OID 16545)
-- Name: idx_student_registrations_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_student_registrations_status ON public.student_registrations USING btree (status);


--
-- TOC entry 4931 (class 1259 OID 16406)
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_users_email ON public.users USING btree (email);


--
-- TOC entry 4973 (class 2606 OID 16546)
-- Name: access_requests access_requests_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.access_requests
    ADD CONSTRAINT access_requests_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- TOC entry 4974 (class 2606 OID 16551)
-- Name: access_requests access_requests_file_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.access_requests
    ADD CONSTRAINT access_requests_file_id_fkey FOREIGN KEY (file_id) REFERENCES public.document_files(id) ON DELETE CASCADE;


--
-- TOC entry 4975 (class 2606 OID 16556)
-- Name: document_files document_files_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.document_files
    ADD CONSTRAINT document_files_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- TOC entry 4976 (class 2606 OID 16561)
-- Name: document_views document_views_document_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.document_views
    ADD CONSTRAINT document_views_document_id_fkey FOREIGN KEY (document_id) REFERENCES public.documents(id) ON DELETE CASCADE;


--
-- TOC entry 4977 (class 2606 OID 16566)
-- Name: documents documents_fakultas_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_fakultas_id_fkey FOREIGN KEY (fakultas_id) REFERENCES public.fakultas(id) ON DELETE SET NULL;


--
-- TOC entry 4978 (class 2606 OID 16571)
-- Name: documents documents_prodi_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_prodi_id_fkey FOREIGN KEY (prodi_id) REFERENCES public.prodi(id) ON DELETE SET NULL;


--
-- TOC entry 4972 (class 2606 OID 16432)
-- Name: prodi prodi_fakultas_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prodi
    ADD CONSTRAINT prodi_fakultas_id_fkey FOREIGN KEY (fakultas_id) REFERENCES public.fakultas(id) ON DELETE RESTRICT;


-- Completed on 2026-04-14 11:11:27

--
-- PostgreSQL database dump complete
--

\unrestrict DKCye47nOgfNvn43YL427Xe9wpOI4p13gqDV1v6ASgPYjFBGRTqf5HmaoLZLpEl

