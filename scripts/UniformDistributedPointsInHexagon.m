function [x,y] = UniformDistributedPointsInHexagon(L,N)
% This function generates N number of uniformly distributed points inside 
% a hexagon of radius L (unit of length).
%
% INPUTS:
%       L = Scalar, real valued variable specifying radius of hexagon in
%       unit of length. Default value of L = 1 (unit of length).
%       N = Scalar, real valued variable specifying number of points to be
%           generated. Default value of N = 1e6
%
% OUTPUTS:
%       x = Vector, real valued variable specifying x coordinate inside
%           hexagon of radius L
%       y = Vector, real valued variable specifying y coordinate inside
%           heaagon of radius L
%
% USAGES:
% [x,y] = UniformDistributedPointsInHexagon; 
% Generates N = 1e6 uniformly distributed points inside a hexagon of
% radius 1 (unit of length).
%
% [x,y] = UniformDistributedPointsInHexagon(L); 
% Generates N = 1e6 uniformly distributed points inside a hexagon of
% radius L (unit of length).
%
% [x,y] = UniformDistributedPointsInHexagon(L,N); 
% Generates N uniformly distributed points inside a hexagon of
% radius L (unit of length).
%
% EXAMPLES:
% %% Release memory, clear screen, and figure
% clear;clc;clf
% 
% %%
% L = 2; % Radius of hexagon in meter
% N = [1e3,1e6]; % Number of random points to be generated
% % Calling function
% [xs1,ys1] = UniformDistributedPointsInHexagon(L,N(1));
% [xs2,ys2] = UniformDistributedPointsInHexagon(L,N(2));
% 
% %% Plotting results
% subplot(121);
% plot(xs1,ys1,'.');
% axis square;
% title('N = 1000')
% subplot(122);
% plot(xs2,ys2,'.');
% axis square;
% title('N = 1000000')
%
% REFERENCES:
% [1] Mouhamed Abdulla and Yousef R. Shayan, "Cellular-based Statistical
% Model for Mobile Dispersion"
% Department of Electrical and Computer Engineering
% Concordia University
% Montr�al, Qu�bec, Canada
% Email: {m_abdull, yshayan}@ece.concordia.ca
%
%
% AUTHOR:
% Ashish (Meet) Meshram
% meetashish85@gmail.com
%
% SEE ALSO:
% rand
% Checking Input Arguments
if nargin<1||isempty(L),L = 1;end
if nargin<2||isempty(N),N = 1e6;end
% Implementation
U = rand(1,N);           % Uniformly Distributed Random Number from 0 to 1
U1 = U(U>0 & U<=1/6);    % Random Number, U in (0,1/6]   
U2 = U(U>1/6 & U<=5/6); % Random Number, U in [1/6,5/6] 
U3 = U(U>5/6 & U<1);    % Random Number, U in [5/6,1)
X1 = L*(sqrt(3*U1/2) - 1);
X2 = (3*L/4)*(2*U2 - 1);
X3 = L*(1 - sqrt((3*(1-U3))/(2)));
x = zeros(1,N);
x((U>0 & U<=1/6)) = X1;
x((U>=1/6 & U<=5/6)) = X2;
x((U>=5/6 & U<1)) = X3;
X1 = x(x > -L & x <= -L/2);     % Random Number X in the range (-L,-L/2]
X2 = x(x >= -L/2 & x <= L/2);   % Random Number X in the range [-L/2,L/2]
X3 = x(x >= L/2 & x < L);       % Random Number X in the range [L/2,L]
Y1 = zeros(size(X1));
for k = 1:length(X1)
    a = -sqrt(3)*(X1(k)+L);
    b = sqrt(3)*(X1(k)+L);
    Y1(k) = a + (b - a)*rand;
end
Y2 = zeros(size(X2));
for k = 1:length(X2)
    a = -sqrt(3)*L/2;
    b = sqrt(3)*L/2;
    Y2(k) = a + (b - a)*rand;
end
Y3 = zeros(size(X3));
for k = 1:length(X3)
    a = -sqrt(3)*(L - X3(k));
    b = sqrt(3)*(L - X3(k));
    Y3(k) = a + (b - a)*rand;
end
y = zeros(size(x));
y(x > -L & x <= -L/2) = Y1;
y(x >= -L/2 & x <= L/2) = Y2;
y(x >= L/2 & x < L) = Y3;