import { NotFound } from "@/components/global/widget/not-found";

export default function NotFoundPage() {
  return (
    <NotFound
      title="404 - Page Not Found"
      message="The page you are looking for has taken an unexpected vacation. Let's get you back on track."
      backTo="/"
      backText="Go to Dashboard"
    />
  );
}
