function drawPolyGon(centre,radius,fmt,linewidth)
if ~exist('fmt')
    fmt='-r';
end
if ~exist('linewidth')
    linewidth=2;
end
for k=1:length(centre)
[x,y]=pol2cart([0:5]*pi/3+pi/6,radius);
x=x+real(centre(k));
y=y+imag(centre(k));

x(end+1)=x(1);
y(end+1)=y(1);


   plot(x,y,fmt,'LineWidth',linewidth) ;hold on;
end