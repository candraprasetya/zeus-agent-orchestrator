import type { SVGProps } from "react";
import { HugeiconsIcon } from "@hugeicons/react";
import { FileExportIcon } from "@hugeicons/core-free-icons";

/** Open a file in the main workspace view, rendered with HugeIcons. */
export function VscodeGoToFileIcon(props: SVGProps<SVGSVGElement>) {
	return <HugeiconsIcon icon={FileExportIcon} {...(props as any)} />;
}
