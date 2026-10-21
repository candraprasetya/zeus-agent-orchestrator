import { HugeiconsIcon } from "@hugeicons/react";
import { Alien02Icon } from "@hugeicons/core-free-icons";

export function AOMascot({ className }: { className: string }) {
	return (
		<span className={`inline-flex items-center justify-center ${className}`} aria-hidden="true">
			<HugeiconsIcon icon={Alien02Icon} size="100%" className="size-full text-current" strokeWidth={1.5} />
		</span>
	);
}
