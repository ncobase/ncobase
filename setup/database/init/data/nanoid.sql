/*
 * ref: https://github.com/viascom/nanoid-postgres
 */

CREATE
  EXTENSION IF NOT EXISTS pgcrypto;

CREATE
  OR REPLACE FUNCTION nanoid(size int DEFAULT 16,
                             alphabet text DEFAULT '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ')
  RETURNS text
  LANGUAGE plpgsql
  stable
AS
$$
DECLARE
  idBuilder text := '';
  i
            int  := 0;
  bytes
            bytea;
  alphabetIndex
            int;
  mask
            int;
  step
            int;
BEGIN
  mask
    := (2 << cast(floor(log(length(alphabet) - 1) / log(2)) as int)) - 1;
  step
    := cast(ceil(1.6 * mask * size / length(alphabet)) AS int);

  while
    true
    loop
      bytes := gen_random_bytes(size);
      while
        i < size
        loop
          alphabetIndex := get_byte(bytes, i) & mask;
          if
            alphabetIndex < length(alphabet) then
            idBuilder := idBuilder || substr(alphabet, alphabetIndex, 1);
            if
              length(idBuilder) = size then
              return idBuilder;
            end if;
          end if;
          i
            = i + 1;
        end loop;
      i
        := 0;
    end loop;
END
$$;
