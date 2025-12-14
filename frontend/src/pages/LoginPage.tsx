import type { ApiResponse } from "@/api/types";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useAxios } from "@/hooks/useAxios";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import { z } from "zod";

const loginSchema = z.object({
  email: z.email("nieprawidłowy email"),
  password: z.string().min(1, "hasło wymagane"),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
  const navigate = useNavigate();

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const axios = useAxios();
  const queryClient = useQueryClient();

  const loginUser = useMutation({
    mutationFn: async (data: { email: string; password: string }) => {
      await axios.post<ApiResponse<void>>("/auth/login", data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["user"] });
      navigate("/");
    },
    onError: () => {
      form.setError("root", {
        message: "nieprawidłowy email lub hasło",
      });
    },
  });

  const onSubmit = async (data: LoginForm) => {
    await loginUser.mutateAsync(data);
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="w-full max-w-xs">
        {/* Header - surowy styl terminala */}
        <div className="mb-8 text-center">
          <div className="font-mono text-primary text-lg mb-2">[RSP]</div>
          <h1 className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
            rozproszony system plików
          </h1>
        </div>

        {/* Form container - bez cieni, prostokątny */}
        <div className="border border-border bg-card">
          <div className="border-b border-border px-4 py-2">
            <span className="text-xs font-mono uppercase tracking-wide text-muted-foreground">
              :: logowanie ::
            </span>
          </div>

          <div className="p-4">
            <Form {...form}>
              <form
                onSubmit={form.handleSubmit(onSubmit)}
                className="space-y-4"
              >
                {form.formState.errors.root && (
                  <div className="p-2 border border-destructive bg-destructive/10 text-destructive text-xs font-mono">
                    {form.formState.errors.root.message}
                  </div>
                )}

                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>email</FormLabel>
                      <FormControl>
                        <Input
                          type="email"
                          placeholder="user@example.com"
                          autoComplete="email"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>hasło</FormLabel>
                      <FormControl>
                        <Input
                          type="password"
                          autoComplete="current-password"
                          {...field}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <Button
                  type="submit"
                  className="w-full"
                  disabled={loginUser.isPending}
                >
                  {loginUser.isPending ? (
                    <>
                      <Loader2 className="w-3.5 h-3.5 animate-spin" />
                      logowanie...
                    </>
                  ) : (
                    "zaloguj"
                  )}
                </Button>
              </form>
            </Form>
          </div>
        </div>

        <p className="mt-4 text-center text-xs text-muted-foreground font-mono">
          brak konta?{" "}
          <Link to="/register" className="text-primary hover:underline">
            zarejestruj się
          </Link>
        </p>
      </div>
    </div>
  );
}
