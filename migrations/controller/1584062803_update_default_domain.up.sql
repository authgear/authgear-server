UPDATE domain SET "assignment" = 'default'
FROM(
  SELECT
    d.id
  FROM
    domain as d
  FULL JOIN
    custom_domain as cd
  ON
    d.domain = cd.domain
  WHERE
    cd.id IS NULL
) AS sub
WHERE domain.id = sub.id;
