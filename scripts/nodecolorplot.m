%load nodeinfo.dat % NodeID,Freq,X,Y,SINR
%nodeinfo =nodeinfo(find(mod(nodeinfo(:,7),2)==0),:);
 
col=5;  % column where SINR values are listed
nodeinfo=nodecolortable;
frequency= unique(nodeinfo(:,2))';
for f=frequency
nodeinfoTable=nodeinfo(find(nodeinfo(:,2)==f),:);

% filtered=1:length(nodeinfoTable)
 filtered=find(nodeinfoTable(:,col)>-7);
% filtered=find(nodeinfoTable(:,col)>-5);
% filtered=find(uelocations(:,6)<3500);



figure

sinr=nodeinfoTable(filtered,col);
cmap=colormap;
LEVELS=length(cmap);
minsinr=-36;
maxsinr=60;
sinrrange=(maxsinr-minsinr);
cedges=[0:LEVELS-1]*sinrrange/LEVELS+(minsinr);

clevel=quantiz(sinr,cedges);

N=length(nodeinfoTable(filtered,1));
 	deltasize=80/14;
	S=80*ones(N,1);
sinrrange
LEVELS
delta = (sinrrange/LEVELS)
C=floor(sinr/delta);
C=cedges(clevel);

scatter3(nodeinfoTable(filtered,3),nodeinfoTable(filtered,4),nodeinfoTable(filtered,col),S,C,'filled')

colorbar
view(2)
title(f)
end
hold on
plot(bslocations(:,2),bslocations(:,3),'*k','MarkerSize',10)
hold on;
plot(antennalocations(:,1),antennalocations(:,2),'Or','MarkerSize',10) 

% stable=stable(1:500,:);
bestbsid=stable(:,7);

drawPolyGon(complex(bslocations(:,2),bslocations(:,3)),ISD/sqrt(3));
drawPolyGon(complex(antennalocations(:,1),antennalocations(:,2)),ISD/sqrt(3),'k',3);

 
