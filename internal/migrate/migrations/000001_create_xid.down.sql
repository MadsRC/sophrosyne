DROP FUNCTION public.xid_counter(_xid public.xid);

DROP FUNCTION public.xid_pid(_xid public.xid);

DROP FUNCTION public.xid_machine(_xid public.xid);

DROP FUNCTION public.xid_time(_xid public.xid);

DROP FUNCTION xid(_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP);

DROP FUNCTION public.xid_decode(_xid public.xid);

DROP FUNCTION public.xid_encode(_id int[]);

DROP FUNCTION public._xid_machine_id();

DROP SEQUENCE public.xid_serial MINVALUE 0 MAXVALUE 16777215 CYCLE;

DROP DOMAIN public.xid;
