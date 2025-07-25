import { useState, useEffect, Dispatch, SetStateAction } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";

export function usePaginatedSearchParams(): {
  offset: number;
  setOffset: Dispatch<SetStateAction<number>>;
  searchKeyword: string;
  setSearchKeyword: Dispatch<SetStateAction<string>>;
} {
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const initialOffset = Number(searchParams.get("offset")) || 0;
  const initialSearchKeyword = searchParams.get("searchKeyword") || "";

  const [offset, setOffset] = useState(initialOffset);
  const [searchKeyword, setSearchKeyword] = useState(initialSearchKeyword);

  useEffect(() => {
    const newParams = new URLSearchParams(searchParams);
    if (offset > 0) {
      newParams.set("offset", String(offset));
    } else {
      newParams.delete("offset");
    }
    if (searchKeyword) {
      newParams.set("searchKeyword", searchKeyword);
    } else {
      newParams.delete("searchKeyword");
    }
    navigate(
      {
        search: newParams.toString(),
        hash: location.hash,
        pathname: location.pathname,
      },
      { replace: true }
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [offset, searchKeyword]);

  return { offset, setOffset, searchKeyword, setSearchKeyword };
}
