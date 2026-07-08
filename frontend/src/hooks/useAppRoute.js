import { useEffect, useState } from "react";

export function useAppRoute() {
  const [route, setRoute] = useState(parseRoute());

  useEffect(() => {
    const onPop = () => setRoute(parseRoute());
    window.addEventListener("popstate", onPop);
    return () => window.removeEventListener("popstate", onPop);
  }, []);

  const navigate = (path) => {
    if (window.location.pathname !== path) {
      window.history.pushState({}, "", path);
    }
    setRoute(parseRoute(path));
  };

  return { route, navigate };
}

function parseRoute(path = window.location.pathname) {
  const teamMatch = path.match(/^\/(?:server\/)?team\/(\d+)\/?$/);
  if (teamMatch) return { name: "team", id: Number(teamMatch[1]) };

  const matchMatch = path.match(/^\/(?:server\/)?match\/(\d+)\/?$/);
  if (matchMatch) return { name: "match", id: Number(matchMatch[1]) };

  return { name: "dashboard" };
}
