import { forwardRef, type SVGProps, type ForwardRefExoticComponent, type RefAttributes } from "react";
import { HugeiconsIcon, type IconSvgElement } from "@hugeicons/react";
import {
  ActivityIcon as _ActivityIcon,
  Alert02Icon as _Alert02Icon,
  AlertCircleIcon as _AlertCircleIcon,
  ArchiveIcon as _ArchiveIcon,
  Attachment01Icon as _Attachment01Icon,
  ArrowDown01Icon as _ArrowDown01Icon,
  ArrowDownIcon as _ArrowDownIcon,
  ArrowDownUp as _ArrowDownUp,
  ArrowLeft01Icon as _ArrowLeft01Icon,
  ArrowLeftIcon as _ArrowLeftIcon,
  ArrowLeftRightIcon as _ArrowLeftRightIcon,
  ArrowRight01Icon as _ArrowRight01Icon,
  ArrowRightIcon as _ArrowRightIcon,
  ArrowTurnBackwardIcon as _ArrowTurnBackwardIcon,
  ArrowTurnDownIcon as _ArrowTurnDownIcon,
  ArrowUp01Icon as _ArrowUp01Icon,
  ArrowUp02Icon as _ArrowUp02Icon,
  ArrowUpIcon as _ArrowUpIcon,
  ArrowUpRightIcon as _ArrowUpRightIcon,
  BadgeCheckIcon as _BadgeCheckIcon,
  BanIcon as _BanIcon,
  BellIcon as _BellIcon,
  BoldIcon as _BoldIcon,
  BookOpenIcon as _BookOpenIcon,
  BotIcon as _BotIcon,
  BoxIcon as _BoxIcon,
  Brain01Icon as _Brain01Icon,
  Bug01Icon as _Bug01Icon,
  Calendar01Icon as _Calendar01Icon,
  Camera01Icon as _Camera01Icon,
  CancelCircleIcon as _CancelCircleIcon,
  CheckCheckIcon as _CheckCheckIcon,
  CheckIcon as _CheckIcon,
  CheckmarkCircle01Icon as _CheckmarkCircle01Icon,
  CircleIcon as _CircleIcon,
  Clock01Icon as _Clock01Icon,
  CloudIcon as _CloudIcon,
  CoffeeIcon as _CoffeeIcon,
  Columns2 as _Columns2,
  Comment01Icon as _Comment01Icon,
  Comment02Icon as _Comment02Icon,
  CommentAdd01Icon as _CommentAdd01Icon,
  ComputerIcon as _ComputerIcon,
  ComputerSettingsIcon as _ComputerSettingsIcon,
  ConsoleIcon as _ConsoleIcon,
  CookieIcon as _CookieIcon,
  CopyIcon as _CopyIcon,
  CreditCardIcon as _CreditCardIcon,
  CursorPointer01Icon as _CursorPointer01Icon,
  DashboardSpeed01Icon as _DashboardSpeed01Icon,
  Delete02Icon as _Delete02Icon,
  DocumentCodeIcon as _DocumentCodeIcon,
  DownloadIcon as _DownloadIcon,
  EraserIcon as _EraserIcon,
  EyeIcon as _EyeIcon,
  EyeOffIcon as _EyeOffIcon,
  File01Icon as _File01Icon,
  FileAddIcon as _FileAddIcon,
  FileCodeIcon as _FileCodeIcon,
  FileIcon as _FileIcon,
  FlashIcon as _FlashIcon,
  FloppyDiskIcon as _FloppyDiskIcon,
  Folder01Icon as _Folder01Icon,
  Folder02Icon as _Folder02Icon,
  FolderAddIcon as _FolderAddIcon,
  FolderCodeIcon as _FolderCodeIcon,
  FolderIcon as _FolderIcon,
  FolderLibraryIcon as _FolderLibraryIcon,
  FullScreenIcon as _FullScreenIcon,
  GitBranchIcon as _GitBranchIcon,
  GitCommitIcon as _GitCommitIcon,
  GitCompareIcon as _GitCompareIcon,
  GitForkIcon as _GitForkIcon,
  GitMergeIcon as _GitMergeIcon,
  GitPullRequestIcon as _GitPullRequestIcon,
  Globe02Icon as _Globe02Icon,
  GlobeIcon as _GlobeIcon,
  GripVerticalIcon as _GripVerticalIcon,
  Heading1 as _Heading1,
  Heading2 as _Heading2,
  Heading3 as _Heading3,
  Heading4 as _Heading4,
  Heading5 as _Heading5,
  HelpCircleIcon as _HelpCircleIcon,
  HistoryIcon as _HistoryIcon,
  Home01Icon as _Home01Icon,
  ImageIcon as _ImageIcon,
  ImportIcon as _ImportIcon,
  InboxIcon as _InboxIcon,
  InfoIcon as _InfoIcon,
  ItalicIcon as _ItalicIcon,
  Key01Icon as _Key01Icon,
  KeyboardIcon as _KeyboardIcon,
  LayoutDashboardIcon as _LayoutDashboardIcon,
  LeftToRightListNumberIcon as _LeftToRightListNumberIcon,
  LightbulbIcon as _LightbulbIcon,
  Link2 as _Link2,
  LinkIcon as _LinkIcon,
  LinkSquare02Icon as _LinkSquare02Icon,
  ListChecksIcon as _ListChecksIcon,
  ListIcon as _ListIcon,
  Loading01Icon as _Loading01Icon,
  Loading03Icon as _Loading03Icon,
  LockIcon as _LockIcon,
  LogInIcon as _LogInIcon,
  LogOutIcon as _LogOutIcon,
  MessageSquareDiffIcon as _MessageSquareDiffIcon,
  MessageSquareQuoteIcon as _MessageSquareQuoteIcon,
  Minimize2 as _Minimize2,
  MinusIcon as _MinusIcon,
  MinusSignCircleIcon as _MinusSignCircleIcon,
  MonitorSmartphoneIcon as _MonitorSmartphoneIcon,
  MoreHorizontalIcon as _MoreHorizontalIcon,
  MoreVerticalIcon as _MoreVerticalIcon,
  NoteIcon as _NoteIcon,
  Notification03Icon as _Notification03Icon,
  PauseIcon as _PauseIcon,
  PencilEdit01Icon as _PencilEdit01Icon,
  PencilEdit02Icon as _PencilEdit02Icon,
  PilcrowIcon as _PilcrowIcon,
  PinIcon as _PinIcon,
  PlayIcon as _PlayIcon,
  PlugIcon as _PlugIcon,
  PlusIcon as _PlusIcon,
  QrCodeIcon as _QrCodeIcon,
  QuoteIcon as _QuoteIcon,
  RefreshIcon as _RefreshIcon,
  RegexIcon as _RegexIcon,
  RepeatIcon as _RepeatIcon,
  RotateLeft01Icon as _RotateLeft01Icon,
  RotateRight01Icon as _RotateRight01Icon,
  Rows3Icon as _Rows3Icon,
  ScrollTextIcon as _ScrollTextIcon,
  SearchIcon as _SearchIcon,
  SearchListIcon as _SearchListIcon,
  SendIcon as _SendIcon,
  ServerIcon as _ServerIcon,
  Settings01Icon as _Settings01Icon,
  Settings02Icon as _Settings02Icon,
  SlidersHorizontalIcon as _SlidersHorizontalIcon,
  SheetIcon as _SheetIcon,
  ShieldAlertIcon as _ShieldAlertIcon,
  ShieldCheckIcon as _ShieldCheckIcon,
  ShieldXIcon as _ShieldXIcon,
  ShuffleIcon as _ShuffleIcon,
  SidebarLeft01Icon as _SidebarLeft01Icon,
  SidebarRight01Icon as _SidebarRight01Icon,
  SmartphoneIcon as _SmartphoneIcon,
  Sorting01Icon as _Sorting01Icon,
  Sorting05Icon as _Sorting05Icon,
  SourceCodeIcon as _SourceCodeIcon,
  SparklesIcon as _SparklesIcon,
  SquareArrowDiagonal01Icon as _SquareArrowDiagonal01Icon,
  SquareIcon as _SquareIcon,
  StarIcon as _StarIcon,
  StrikethroughIcon as _StrikethroughIcon,
  Sun01Icon as _Sun01Icon,
  Tablet01Icon as _Tablet01Icon,
  Tag01Icon as _Tag01Icon,
  TerminalIcon as _TerminalIcon,
  TextFontIcon as _TextFontIcon,
  TextWrapIcon as _TextWrapIcon,
  Undo2 as _Undo2,
  UserIcon as _UserIcon,
  UsersIcon as _UsersIcon,
  VideoIcon as _VideoIcon,
  WrenchIcon as _WrenchIcon,
  XIcon as _XIcon
} from "@hugeicons/core-free-icons";

export interface LucideProps extends SVGProps<SVGSVGElement> {
	size?: number | string;
	color?: string;
	strokeWidth?: number | string;
	absoluteStrokeWidth?: boolean;
}

export type LucideIcon = ForwardRefExoticComponent<LucideProps & RefAttributes<SVGSVGElement>>;

export function createHugeIcon(kebabName: string, iconData: IconSvgElement): LucideIcon {
	const Comp = forwardRef<SVGSVGElement, LucideProps>(({
		size = 24,
		color = "currentColor",
		strokeWidth = 2,
		className,
		children: _children,
		...rest
	}, ref) => {
		const classes = ["lucide", `lucide-${kebabName}`, "hugeicons", `hugeicons-${kebabName}`, className]
			.filter(Boolean)
			.join(" ");

		return (
			<HugeiconsIcon
				ref={ref}
				icon={iconData}
				size={size}
				color={color}
				strokeWidth={Number(strokeWidth) || 2}
				className={classes}
				{...(rest as any)}
			/>
		);
	});
	Comp.displayName = kebabName;
	return Comp;
}

export const Activity = createHugeIcon("activity", _ActivityIcon);
export const AlertTriangle = createHugeIcon("alert-triangle", _Alert02Icon);
export const Archive = createHugeIcon("archive", _ArchiveIcon);
export const ArrowDown = createHugeIcon("arrow-down", _ArrowDownIcon);
export const ArrowDownUp = createHugeIcon("arrow-down-up", _ArrowDownUp);
export const ArrowLeft = createHugeIcon("arrow-left", _ArrowLeftIcon);
export const ArrowRight = createHugeIcon("arrow-right", _ArrowRightIcon);
export const ArrowRightLeft = createHugeIcon("arrow-right-left", _ArrowLeftRightIcon);
export const ArrowUp = createHugeIcon("arrow-up", _ArrowUpIcon);
export const ArrowUpFromLine = createHugeIcon("arrow-up-from-line", _ArrowUp02Icon);
export const ArrowUpRight = createHugeIcon("arrow-up-right", _ArrowUpRightIcon);
export const BadgeCheck = createHugeIcon("badge-check", _BadgeCheckIcon);
export const Ban = createHugeIcon("ban", _BanIcon);
export const Bell = createHugeIcon("bell", _BellIcon);
export const BellRing = createHugeIcon("bell-ring", _Notification03Icon);
export const Bold = createHugeIcon("bold", _BoldIcon);
export const BookOpen = createHugeIcon("book-open", _BookOpenIcon);
export const Bot = createHugeIcon("bot", _BotIcon);
export const Box = createHugeIcon("box", _BoxIcon);
export const Brain = createHugeIcon("brain", _Brain01Icon);
export const Bug = createHugeIcon("bug", _Bug01Icon);
export const CalendarClock = createHugeIcon("calendar-clock", _Calendar01Icon);
export const Camera = createHugeIcon("camera", _Camera01Icon);
export const CaseSensitive = createHugeIcon("case-sensitive", _TextFontIcon);
export const Check = createHugeIcon("check", _CheckIcon);
export const CheckCheck = createHugeIcon("check-check", _CheckCheckIcon);
export const CheckCircle2 = createHugeIcon("check-circle-2", _CheckmarkCircle01Icon);
export const ChevronDown = createHugeIcon("chevron-down", _ArrowDown01Icon);
export const ChevronDownIcon = createHugeIcon("chevron-down", _ArrowDown01Icon);
export const ChevronLeft = createHugeIcon("chevron-left", _ArrowLeft01Icon);
export const ChevronRight = createHugeIcon("chevron-right", _ArrowRight01Icon);
export const ChevronRightIcon = createHugeIcon("chevron-right", _ArrowRight01Icon);
export const ChevronUp = createHugeIcon("chevron-up", _ArrowUp01Icon);
export const ChevronUpIcon = createHugeIcon("chevron-up", _ArrowUp01Icon);
export const ChevronsDownUp = createHugeIcon("chevrons-down-up", _Sorting05Icon);
export const ChevronsUpDown = createHugeIcon("chevrons-up-down", _Sorting01Icon);
export const ListChevronsDownUp = ChevronsDownUp;
export const ListChevronsUpDown = ChevronsUpDown;
export const Circle = createHugeIcon("circle", _CircleIcon);
export const CircleAlert = createHugeIcon("circle-alert", _AlertCircleIcon);
export const CircleArrowUp = createHugeIcon("circle-arrow-up", _ArrowUp01Icon);
export const CircleCheck = createHugeIcon("circle-check", _CheckmarkCircle01Icon);
export const CircleDashed = createHugeIcon("circle-dashed", _CircleIcon);
export const CircleHelp = createHugeIcon("circle-help", _HelpCircleIcon);
export const CircleMinus = createHugeIcon("circle-minus", _MinusSignCircleIcon);
export const Clock = createHugeIcon("clock", _Clock01Icon);
export const Clock3 = createHugeIcon("clock-3", _Clock01Icon);
export const Cloud = createHugeIcon("cloud", _CloudIcon);
export const Code2 = createHugeIcon("code-2", _SourceCodeIcon);
export const Coffee = createHugeIcon("coffee", _CoffeeIcon);
export const Columns2 = createHugeIcon("columns-2", _Columns2);
export const Cookie = createHugeIcon("cookie", _CookieIcon);
export const Copy = createHugeIcon("copy", _CopyIcon);
export const CornerDownLeft = createHugeIcon("corner-down-left", _ArrowTurnBackwardIcon);
export const CornerDownRight = createHugeIcon("corner-down-right", _ArrowTurnDownIcon);
export const CornerUpRight = createHugeIcon("corner-up-right", _ArrowTurnDownIcon);
export const CreditCard = createHugeIcon("credit-card", _CreditCardIcon);
export const Download = createHugeIcon("download", _DownloadIcon);
export const Eraser = createHugeIcon("eraser", _EraserIcon);
export const ExternalLink = createHugeIcon("external-link", _LinkSquare02Icon);
export const Eye = createHugeIcon("eye", _EyeIcon);
export const EyeOff = createHugeIcon("eye-off", _EyeOffIcon);
export const File = createHugeIcon("file", _FileIcon);
export const FileCode2 = createHugeIcon("file-code-2", _FileCodeIcon);
export const FileDiff = createHugeIcon("file-diff", _FileCodeIcon);
export const FileJson2 = createHugeIcon("file-json-2", _DocumentCodeIcon);
export const FilePlus = createHugeIcon("file-plus", _FileAddIcon);
export const FileText = createHugeIcon("file-text", _File01Icon);
export const Folder = createHugeIcon("folder", _FolderIcon);
export const FolderClosed = createHugeIcon("folder-closed", _Folder01Icon);
export const FolderGit2 = createHugeIcon("folder-git-2", _FolderCodeIcon);
export const FolderOpen = createHugeIcon("folder-open", _Folder02Icon);
export const FolderPlus = createHugeIcon("folder-plus", _FolderAddIcon);
export const FolderTree = createHugeIcon("folder-tree", _FolderLibraryIcon);
export const Folders = createHugeIcon("folders", _FolderLibraryIcon);
export const Gauge = createHugeIcon("gauge", _DashboardSpeed01Icon);
export const GitBranch = createHugeIcon("git-branch", _GitBranchIcon);
export const GitCommitHorizontal = createHugeIcon("git-commit-horizontal", _GitCommitIcon);
export const GitCompareArrows = createHugeIcon("git-compare-arrows", _GitCompareIcon);
export const GitFork = createHugeIcon("git-fork", _GitForkIcon);
export const GitMerge = createHugeIcon("git-merge", _GitMergeIcon);
export const GitPullRequest = createHugeIcon("git-pull-request", _GitPullRequestIcon);
export const GitPullRequestArrow = createHugeIcon("git-pull-request-arrow", _GitPullRequestIcon);
export const GitPullRequestClosed = createHugeIcon("git-pull-request-closed", _GitPullRequestIcon);
export const GitPullRequestDraft = createHugeIcon("git-pull-request-draft", _GitPullRequestIcon);
export const Globe = createHugeIcon("globe", _GlobeIcon);
export const Globe2 = createHugeIcon("globe-2", _Globe02Icon);
export const GripVertical = createHugeIcon("grip-vertical", _GripVerticalIcon);
export const GripVerticalIcon = createHugeIcon("grip-vertical", _GripVerticalIcon);
export const Hand = createHugeIcon("hand", _CursorPointer01Icon);
export const Heading1 = createHugeIcon("heading-1", _Heading1);
export const Heading2 = createHugeIcon("heading-2", _Heading2);
export const Heading3 = createHugeIcon("heading-3", _Heading3);
export const Heading4 = createHugeIcon("heading-4", _Heading4);
export const Heading5 = createHugeIcon("heading-5", _Heading5);
export const History = createHugeIcon("history", _HistoryIcon);
export const House = createHugeIcon("house", _Home01Icon);
export const Image = createHugeIcon("image", _ImageIcon);
export const Import = createHugeIcon("import", _ImportIcon);
export const Inbox = createHugeIcon("inbox", _InboxIcon);
export const Info = createHugeIcon("info", _InfoIcon);
export const Italic = createHugeIcon("italic", _ItalicIcon);
export const KeyRound = createHugeIcon("key-round", _Key01Icon);
export const Keyboard = createHugeIcon("keyboard", _KeyboardIcon);
export const LayoutDashboard = createHugeIcon("layout-dashboard", _LayoutDashboardIcon);
export const Lightbulb = createHugeIcon("lightbulb", _LightbulbIcon);
export const Link = createHugeIcon("link", _LinkIcon);
export const Link2 = createHugeIcon("link-2", _Link2);
export const List = createHugeIcon("list", _ListIcon);
export const ListChecks = createHugeIcon("list-checks", _ListChecksIcon);
export const ListOrdered = createHugeIcon("list-ordered", _LeftToRightListNumberIcon);
export const ListTodo = createHugeIcon("list-todo", _ListChecksIcon);
export const Loader2 = createHugeIcon("loader-2", _Loading01Icon);
export const LoaderCircle = createHugeIcon("loader-circle", _Loading03Icon);
export const Lock = createHugeIcon("lock", _LockIcon);
export const LogIn = createHugeIcon("log-in", _LogInIcon);
export const LogOut = createHugeIcon("log-out", _LogOutIcon);
export const Maximize2 = createHugeIcon("maximize-2", _FullScreenIcon);
export const MessageSquare = createHugeIcon("message-square", _Comment01Icon);
export const MessageSquareDiff = createHugeIcon("message-square-diff", _MessageSquareDiffIcon);
export const MessageSquareDot = createHugeIcon("message-square-dot", _Comment02Icon);
export const MessageSquarePlus = createHugeIcon("message-square-plus", _CommentAdd01Icon);
export const MessageSquareQuote = createHugeIcon("message-square-quote", _MessageSquareQuoteIcon);
export const Minimize2 = createHugeIcon("minimize-2", _Minimize2);
export const Minus = createHugeIcon("minus", _MinusIcon);
export const Monitor = createHugeIcon("monitor", _ComputerIcon);
export const MonitorCog = createHugeIcon("monitor-cog", _ComputerSettingsIcon);
export const MonitorSmartphone = createHugeIcon("monitor-smartphone", _MonitorSmartphoneIcon);
export const MoreHorizontal = createHugeIcon("more-horizontal", _MoreHorizontalIcon);
export const MoreVertical = createHugeIcon("more-vertical", _MoreVerticalIcon);
export const MousePointer2 = createHugeIcon("mouse-pointer-2", _CursorPointer01Icon);
export const PanelLeft = createHugeIcon("panel-left", _SidebarLeft01Icon);
export const PanelLeftIcon = createHugeIcon("panel-left", _SidebarLeft01Icon);
export const PanelRight = createHugeIcon("panel-right", _SidebarRight01Icon);
export const Paperclip = createHugeIcon("paperclip", _Attachment01Icon);
export const Pause = createHugeIcon("pause", _PauseIcon);
export const Pencil = createHugeIcon("pencil", _PencilEdit02Icon);
export const PencilLine = createHugeIcon("pencil-line", _PencilEdit02Icon);
export const Pilcrow = createHugeIcon("pilcrow", _PilcrowIcon);
export const Pin = createHugeIcon("pin", _PinIcon);
export const PinOff = createHugeIcon("pin-off", _PinIcon);
export const Play = createHugeIcon("play", _PlayIcon);
export const Plug = createHugeIcon("plug", _PlugIcon);
export const Plus = createHugeIcon("plus", _PlusIcon);
export const QrCode = createHugeIcon("qr-code", _QrCodeIcon);
export const Quote = createHugeIcon("quote", _QuoteIcon);
export const RefreshCw = createHugeIcon("refresh-cw", _RefreshIcon);
export const Regex = createHugeIcon("regex", _RegexIcon);
export const Repeat2 = createHugeIcon("repeat-2", _RepeatIcon);
export const RotateCcw = createHugeIcon("rotate-ccw", _RotateLeft01Icon);
export const RotateCw = createHugeIcon("rotate-cw", _RotateRight01Icon);
export const Rows3 = createHugeIcon("rows-3", _Rows3Icon);
export const Save = createHugeIcon("save", _FloppyDiskIcon);
export const ScrollText = createHugeIcon("scroll-text", _ScrollTextIcon);
export const Search = createHugeIcon("search", _SearchIcon);
export const SearchCheck = createHugeIcon("search-check", _SearchListIcon);
export const SearchX = createHugeIcon("search-x", _SearchIcon);
export const Send = createHugeIcon("send", _SendIcon);
export const Server = createHugeIcon("server", _ServerIcon);
export const Settings = createHugeIcon("settings", _Settings01Icon);
export const Settings2 = createHugeIcon("settings-2", _Settings02Icon);
export const Sheet = createHugeIcon("sheet", _SheetIcon);
export const ShieldAlert = createHugeIcon("shield-alert", _ShieldAlertIcon);
export const ShieldCheck = createHugeIcon("shield-check", _ShieldCheckIcon);
export const ShieldQuestion = createHugeIcon("shield-question", _ShieldAlertIcon);
export const ShieldX = createHugeIcon("shield-x", _ShieldXIcon);
export const Shuffle = createHugeIcon("shuffle", _ShuffleIcon);
export const SlidersHorizontal = createHugeIcon("sliders-horizontal", _SlidersHorizontalIcon);
export const Smartphone = createHugeIcon("smartphone", _SmartphoneIcon);
export const Sparkles = createHugeIcon("sparkles", _SparklesIcon);
export const Square = createHugeIcon("square", _SquareIcon);
export const SquareArrowOutUpRight = createHugeIcon("square-arrow-out-up-right", _SquareArrowDiagonal01Icon);
export const SquarePen = createHugeIcon("square-pen", _PencilEdit01Icon);
export const SquareTerminal = createHugeIcon("square-terminal", _ConsoleIcon);
export const Star = createHugeIcon("star", _StarIcon);
export const StickyNote = createHugeIcon("sticky-note", _NoteIcon);
export const Strikethrough = createHugeIcon("strikethrough", _StrikethroughIcon);
export const Sun = createHugeIcon("sun", _Sun01Icon);
export const Tablet = createHugeIcon("tablet", _Tablet01Icon);
export const Tag = createHugeIcon("tag", _Tag01Icon);
export const Terminal = createHugeIcon("terminal", _TerminalIcon);
export const TerminalSquare = createHugeIcon("terminal-square", _ConsoleIcon);
export const Trash2 = createHugeIcon("trash-2", _Delete02Icon);
export const TriangleAlert = createHugeIcon("triangle-alert", _Alert02Icon);
export const Undo2 = createHugeIcon("undo-2", _Undo2);
export const User = createHugeIcon("user", _UserIcon);
export const UserRound = createHugeIcon("user-round", _UserIcon);
export const Users = createHugeIcon("users", _UsersIcon);
export const Video = createHugeIcon("video", _VideoIcon);
export const WrapText = createHugeIcon("wrap-text", _TextWrapIcon);
export const Wrench = createHugeIcon("wrench", _WrenchIcon);
export const X = createHugeIcon("x", _XIcon);
export const XCircle = createHugeIcon("xcircle", _CancelCircleIcon);
export const XIcon = createHugeIcon("x", _XIcon);
export const Zap = createHugeIcon("zap", _FlashIcon);
