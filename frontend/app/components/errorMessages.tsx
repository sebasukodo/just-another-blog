export function ErrorMessages({
  errors,
}: {
  errors: Record<string, string[]> | null;
}) {
  if (!errors) return null;

  return (
    <ul className="error-messages">
      {Object.entries(errors).flatMap(([field, messages]) =>
        messages.map((msg) => (
          <li key={`${field}-${msg}`}>
            {field} {msg}
          </li>
        )),
      )}
    </ul>
  );
}
