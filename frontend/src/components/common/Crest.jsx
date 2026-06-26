export function Crest({ src, fallback }) {
  const label = fallback || "WC";

  if (!src) {
    return <span className="crest-fallback">{label}</span>;
  }

  return (
    <img
      className="crest"
      src={src}
      alt={label}
      onError={(event) => {
        event.currentTarget.replaceWith(
          Object.assign(document.createElement("span"), {
            className: "crest-fallback",
            textContent: label,
          }),
        );
      }}
    />
  );
}
