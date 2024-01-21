.text
.globl	main
main:
	endbr64
	pushq	%rbp
	movq	%rsp, %rbp

	pushq	$2
	movq 	-8(%rbp), %rax	// move the value previously pushed into rax

	movq	%rbp, %rsp		// fix rsp offset
	popq	%rbp
	ret
