import { cn } from "@/lib/utils";
import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

const badgeVariants = cva(
  "inline-flex items-center border px-2 py-0.5 text-xs font-mono uppercase tracking-wide",
  {
    variants: {
      variant: {
        default: "border-primary bg-primary/20 text-primary",
        secondary: "border-muted-foreground/50 bg-muted text-muted-foreground",
        destructive: "border-destructive bg-destructive/20 text-destructive",
        outline: "border-border text-foreground bg-transparent",
        success: "border-success bg-success/20 text-success",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps
  extends
    React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <div className={cn(badgeVariants({ variant }), className)} {...props} />
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export { Badge, badgeVariants };
